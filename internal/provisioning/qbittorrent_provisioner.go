package provisioning

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/woliveiras/corsarr/internal/credentials"
)

const (
	qbittorrentApplicationID = "qbittorrent"
	qbittorrentUsername      = "corsarr"
	qbittorrentBootstrapUser = "admin"
)

var qbittorrentTemporaryPasswordPattern = regexp.MustCompile(
	`(?i)temporary password is provided for this session:\s*([^\s]+)`,
)

type RuntimeLogReader interface {
	Logs(ctx context.Context, applicationID string, tail int) (string, error)
}

type QBittorrentSession struct {
	client  *http.Client
	baseURL *url.URL
}

type QBittorrentAPI interface {
	Login(
		ctx context.Context,
		username string,
		password credentials.Secret,
	) (*QBittorrentSession, error)
	SetCredentials(
		ctx context.Context,
		session *QBittorrentSession,
		username string,
		password credentials.Secret,
	) error
	EnsureCategories(ctx context.Context, session *QBittorrentSession) error
	EnsureDownloadPaths(ctx context.Context, session *QBittorrentSession) error
}

type QBittorrentProvisioner struct {
	logs             RuntimeLogReader
	credentials      credentials.Store
	client           QBittorrentAPI
	generatePassword func() (credentials.Secret, error)
}

func NewQBittorrentProvisioner(
	logs RuntimeLogReader,
	credentialStore credentials.Store,
	client QBittorrentAPI,
) *QBittorrentProvisioner {
	return &QBittorrentProvisioner{
		logs:             logs,
		credentials:      credentialStore,
		client:           client,
		generatePassword: generateQBittorrentPassword,
	}
}

func (p *QBittorrentProvisioner) Provision(
	ctx context.Context,
	_ string,
	applicationID string,
	_ []string,
) error {
	if applicationID != qbittorrentApplicationID {
		return nil
	}

	storedPassword, err := p.credentials.Load(ctx, credentials.KeyQBitTorrentPassword)
	if err == nil {
		session, loginErr := p.client.Login(ctx, qbittorrentUsername, storedPassword)
		if loginErr != nil {
			return fmt.Errorf("authenticate qBittorrent with stored credential: %w", loginErr)
		}
		if err := p.client.EnsureDownloadPaths(ctx, session); err != nil {
			return fmt.Errorf("ensure qBittorrent download paths: %w", err)
		}
		if err := p.client.EnsureCategories(ctx, session); err != nil {
			return fmt.Errorf("ensure qBittorrent categories: %w", err)
		}
		return nil
	}
	if !errors.Is(err, credentials.ErrCredentialNotFound) {
		return fmt.Errorf("load qBittorrent credential: %w", err)
	}

	logs, err := p.logs.Logs(ctx, qbittorrentApplicationID, 200)
	if err != nil {
		return fmt.Errorf("read qBittorrent bootstrap logs: %w", err)
	}
	temporaryPassword, err := parseQBittorrentTemporaryPassword(logs)
	if err != nil {
		return err
	}
	bootstrapSession, err := p.client.Login(ctx, qbittorrentBootstrapUser, temporaryPassword)
	if err != nil {
		return fmt.Errorf("authenticate qBittorrent bootstrap session: %w", err)
	}

	permanentPassword, err := p.generatePassword()
	if err != nil {
		return fmt.Errorf("generate qBittorrent credential: %w", err)
	}
	if err := p.credentials.Save(
		ctx,
		credentials.KeyQBitTorrentPassword,
		permanentPassword,
	); err != nil {
		return fmt.Errorf("store qBittorrent credential before activation: %w", err)
	}
	if err := p.client.SetCredentials(
		ctx,
		bootstrapSession,
		qbittorrentUsername,
		permanentPassword,
	); err != nil {
		cleanupErr := p.credentials.Delete(ctx, credentials.KeyQBitTorrentPassword)
		changeErr := fmt.Errorf("activate qBittorrent credential: %w", err)
		if cleanupErr != nil {
			return errors.Join(changeErr, fmt.Errorf("remove inactive credential: %w", cleanupErr))
		}
		return changeErr
	}

	verifiedSession, err := p.client.Login(ctx, qbittorrentUsername, permanentPassword)
	if err != nil {
		return fmt.Errorf("verify permanent qBittorrent credential: %w", err)
	}
	if err := p.client.EnsureDownloadPaths(ctx, verifiedSession); err != nil {
		return fmt.Errorf("ensure qBittorrent download paths: %w", err)
	}
	if err := p.client.EnsureCategories(ctx, verifiedSession); err != nil {
		return fmt.Errorf("ensure qBittorrent categories: %w", err)
	}
	return nil
}

func parseQBittorrentTemporaryPassword(logs string) (credentials.Secret, error) {
	matches := qbittorrentTemporaryPasswordPattern.FindStringSubmatch(logs)
	if len(matches) != 2 {
		return credentials.Secret{}, fmt.Errorf("qBittorrent temporary credential was not found")
	}
	password := strings.TrimSpace(matches[1])
	if len(password) < 8 || len(password) > 128 {
		return credentials.Secret{}, fmt.Errorf("qBittorrent temporary credential is invalid")
	}
	return credentials.NewSecret(password), nil
}

func generateQBittorrentPassword() (credentials.Secret, error) {
	random := make([]byte, 18)
	if _, err := rand.Read(random); err != nil {
		return credentials.Secret{}, err
	}
	return credentials.NewSecret(base64.RawURLEncoding.EncodeToString(random)), nil
}
