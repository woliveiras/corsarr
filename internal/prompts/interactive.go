package prompts

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/woliveiras/corsarr/internal/i18n"
	"github.com/woliveiras/corsarr/internal/services"
)

// AskVPN prompts the user if they want to use VPN
func AskVPN(t *i18n.I18n) (bool, error) {
	var useVPN bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(t.T("prompts.vpn_question")).
				Value(&useVPN),
		),
	)

	if err := form.Run(); err != nil {
		return false, err
	}

	return useVPN, nil
}

// SelectServices prompts the user to select which services to use
func SelectServices(t *i18n.I18n, registry *services.Registry, vpnEnabled bool) ([]string, error) {
	// Filter services by VPN compatibility
	availableServices := registry.FilterByVPNCompatibility(vpnEnabled)

	// Build options
	var options []huh.Option[string]
	
	// Group services by category for better organization
	servicesByCategory := make(map[services.ServiceCategory][]*services.Service)
	for _, service := range availableServices {
		servicesByCategory[service.Category] = append(servicesByCategory[service.Category], service)
	}

	// Add services organized by category
	for _, category := range services.AllCategories() {
		servicesInCategory := servicesByCategory[category]
		if len(servicesInCategory) == 0 {
			continue
		}

		// Add services in this category
		for _, service := range servicesInCategory {
			// Get translated description
			description := t.T(service.GetDescriptionKey())
			displayName := fmt.Sprintf("%s (%s)", service.Name, description)
			
			if service.RequiresVPN {
				displayName += t.T("prompts.requires_vpn_suffix")
			}
			if len(service.Dependencies) > 0 {
				displayName += t.T("prompts.has_dependencies_suffix")
			}

			options = append(options, huh.NewOption(displayName, service.ID))
		}
	}

	var selectedIDs []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(t.T("prompts.service_selection")).
				Options(options...).
				Value(&selectedIDs).
				Height(15),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	if len(selectedIDs) == 0 {
		return nil, fmt.Errorf("%s", t.T("errors.no_services_selected"))
	}

	return selectedIDs, nil
}

// AskBasePath prompts for the base path (ARRPATH)
func AskBasePath(t *i18n.I18n, defaultPath string) (string, error) {
	var path string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(t.T("prompts.base_path")).
				Value(&path).
				Placeholder(defaultPath).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("path is required")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	if path == "" {
		path = defaultPath
	}

	return path, nil
}

// AskOutputDirectory prompts for an output directory and optional reuse for volumes.
func AskOutputDirectory(t *i18n.I18n, defaultDir string) (string, bool, bool, error) {
	var dir string
	useSame := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("%s (default: %s)", t.T("prompts.output_directory"), defaultDir)).
				Value(&dir).
				Placeholder(defaultDir),
			huh.NewConfirm().
				Title(t.T("prompts.use_same_directory")).
				Value(&useSame),
		),
	)

	if err := form.Run(); err != nil {
		return "", false, false, err
	}

	cleaned := strings.TrimSpace(strings.ReplaceAll(dir, "\r", ""))
	provided := cleaned != ""
	if !provided {
		cleaned = defaultDir
	}

	return cleaned, useSame, provided, nil
}

// AskTimezone prompts for timezone
func AskTimezone(t *i18n.I18n, defaultTZ string) (string, error) {
	var tz string

	form2 := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(t.T("prompts.timezone")).
				Value(&tz).
				Placeholder(defaultTZ).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("timezone is required")
					}
					return nil
				}),
		),
	)

	if err := form2.Run(); err != nil {
		return "", err
	}

	if tz == "" {
		tz = defaultTZ
	}

	return tz, nil
}

// AskUserIDs prompts for PUID, PGID, and UMASK
func AskUserIDs(t *i18n.I18n) (puid, pgid, umask string, err error) {
	puid = "1000"
	pgid = "1000"
	umask = "002"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(t.T("prompts.puid")).
				Value(&puid).
				Placeholder("1000"),
			huh.NewInput().
				Title(t.T("prompts.pgid")).
				Value(&pgid).
				Placeholder("1000"),
			huh.NewInput().
				Title(t.T("prompts.umask")).
				Value(&umask).
				Placeholder("002"),
		),
	)

	if err := form.Run(); err != nil {
		return "", "", "", err
	}

	return puid, pgid, umask, nil
}

// ConfirmGeneration asks for final confirmation before generating files
func ConfirmGeneration(t *i18n.I18n) (bool, error) {
	var confirm bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(t.T("prompts.confirm_generation")).
				Value(&confirm).
				Affirmative("Yes").
				Negative("No"),
		),
	)

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirm, nil
}
