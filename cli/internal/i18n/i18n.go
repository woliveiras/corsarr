package i18n

import (
	"embed"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var LocaleFS embed.FS

// I18n handles internationalization
type I18n struct {
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	language  string
}

// New creates a new I18n instance
func New(lang string) (*I18n, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Load all supported languages
	supportedLanguages := []string{"en", "pt-br", "es"}
	for _, locale := range supportedLanguages {
		data, err := LocaleFS.ReadFile(fmt.Sprintf("locales/%s.yaml", locale))
		if err != nil {
			return nil, fmt.Errorf("failed to read locale file %s: %w", locale, err)
		}

		if _, err := bundle.ParseMessageFileBytes(data, fmt.Sprintf("%s.yaml", locale)); err != nil {
			return nil, fmt.Errorf("failed to parse locale file %s: %w", locale, err)
		}
	}

	localizer := i18n.NewLocalizer(bundle, lang)

	return &I18n{
		bundle:    bundle,
		localizer: localizer,
		language:  lang,
	}, nil
}

// T translates a message key
func (i *I18n) T(key string, data ...map[string]interface{}) string {
	config := &i18n.LocalizeConfig{
		MessageID: key,
	}

	if len(data) > 0 {
		config.TemplateData = data[0]
	}

	msg, err := i.localizer.Localize(config)
	if err != nil {
		// Fallback to key if translation doesn't exist
		return key
	}
	return msg
}

// Tf translates a message key with formatting (printf style)
func (i *I18n) Tf(key string, args ...interface{}) string {
	msg := i.T(key)
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// GetLanguage returns the current language code
func (i *I18n) GetLanguage() string {
	return i.language
}

// GetLanguageName returns the full language name
func (i *I18n) GetLanguageName() string {
	for _, lang := range SupportedLanguages {
		if lang.Code == i.language {
			return lang.Name
		}
	}
	return i.language
}
