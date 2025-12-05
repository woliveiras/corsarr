package prompts

import (
	"github.com/charmbracelet/huh"
	"github.com/woliveiras/corsarr/internal/generator"
	"github.com/woliveiras/corsarr/internal/i18n"
)

// ConfigureVPN prompts for VPN configuration if VPN is enabled
func ConfigureVPN(translator *i18n.I18n) (*generator.VPNConfig, error) {
	config := &generator.VPNConfig{
		ServiceProvider: "custom",
		Type:            "wireguard",
		PortForwarding:  "off",
		DNSAddress:      "1.1.1.1",
	}

	// VPN type selection
	form1 := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(translator.T("prompts.vpn_type")).
				Options(
					huh.NewOption("WireGuard", "wireguard"),
					huh.NewOption("OpenVPN", "openvpn"),
				).
				Value(&config.Type),
		),
	)

	if err := form1.Run(); err != nil {
		return nil, err
	}

	// Provider and WireGuard config
	if config.Type == "wireguard" {
		form2 := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(translator.T("prompts.vpn_provider")).
					Value(&config.ServiceProvider).
					Placeholder("custom"),
				huh.NewInput().
					Title(translator.T("prompts.vpn_wireguard_private_key")).
					Value(&config.WireguardPrivateKey).
					Password(true).
					EchoMode(huh.EchoModePassword),
				huh.NewInput().
					Title(translator.T("prompts.vpn_wireguard_addresses")).
					Value(&config.WireguardAddresses).
					Placeholder("10.0.0.2/32"),
				huh.NewInput().
					Title(translator.T("prompts.vpn_wireguard_public_key")).
					Value(&config.WireguardPublicKey).
					Placeholder("server public key"),
			),
		)

		if err := form2.Run(); err != nil {
			return nil, err
		}
	} else {
		// OpenVPN provider only
		form2 := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(translator.T("prompts.vpn_provider")).
					Value(&config.ServiceProvider).
					Placeholder("custom"),
			),
		)

		if err := form2.Run(); err != nil {
			return nil, err
		}
	}

	// Port forwarding and DNS
	var enablePortForwarding bool
	form3 := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(translator.T("prompts.vpn_port_forwarding")).
				Value(&enablePortForwarding),
			huh.NewInput().
				Title(translator.T("prompts.vpn_dns")).
				Value(&config.DNSAddress).
				Placeholder("1.1.1.1"),
		),
	)

	if err := form3.Run(); err != nil {
		return nil, err
	}

	if enablePortForwarding {
		config.PortForwarding = "on"
	} else {
		config.PortForwarding = "off"
	}

	return config, nil
}

// ConfigureEnvironment prompts for all environment variables
func ConfigureEnvironment(translator *i18n.I18n, vpnEnabled bool) (*generator.EnvConfig, error) {
	config := generator.NewDefaultEnvConfig()

	// Project name
	projectName, err := AskProjectName(translator, config.ComposeProjectName)
	if err != nil {
		return nil, err
	}
	config.ComposeProjectName = projectName

	// Base path
	basePath, err := AskBasePath(translator, config.ARRPath)
	if err != nil {
		return nil, err
	}
	config.ARRPath = basePath

	// Timezone
	tz, err := AskTimezone(translator, config.Timezone)
	if err != nil {
		return nil, err
	}
	config.Timezone = tz

	// User IDs
	puid, pgid, umask, err := AskUserIDs(translator)
	if err != nil {
		return nil, err
	}
	config.PUID = puid
	config.PGID = pgid
	config.UMASK = umask

	// VPN configuration
	if vpnEnabled {
		vpnConfig, err := ConfigureVPN(translator)
		if err != nil {
			return nil, err
		}
		config.VPNConfig = vpnConfig
	}

	return config, nil
}

// AskProjectName prompts for the compose project name
func AskProjectName(translator *i18n.I18n, defaultName string) (string, error) {
	var projectName string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(translator.T("prompts.project_name")).
				Value(&projectName).
				Placeholder(defaultName).
				Validate(func(s string) error {
					if s == "" {
						projectName = defaultName
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	if projectName == "" {
		projectName = defaultName
	}

	return projectName, nil
}
