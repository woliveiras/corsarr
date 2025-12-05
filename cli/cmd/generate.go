package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/woliveiras/corsarr/internal/generator"
	"github.com/woliveiras/corsarr/internal/i18n"
	"github.com/woliveiras/corsarr/internal/profile"
	"github.com/woliveiras/corsarr/internal/prompts"
	"github.com/woliveiras/corsarr/internal/services"
	"github.com/woliveiras/corsarr/internal/validator"
)

var (
	profileName     string
	outputDir       string
	noInteractive   bool
	useVPN          bool
	dryRun          bool
	saveProfile     bool
	saveProfileName string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate docker-compose.yml and .env files",
	Long: `Generate docker-compose.yml and .env files based on your service selection.

This command will guide you through an interactive process to:
1. Choose whether to use VPN (Gluetun)
2. Select the services you want to use
3. Configure environment variables
4. Generate the files

You can also use a saved profile or run in non-interactive mode.`,
	Run: func(cmd *cobra.Command, args []string) {
		t := GetTranslator()
		
		if err := runGenerate(t); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: %v\n", t.T("errors.generation_failed"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Flags for generate command
	generateCmd.Flags().StringVarP(&profileName, "profile", "p", "", "Load configuration from a saved profile")
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	generateCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Run in non-interactive mode (requires config file or profile)")
	generateCmd.Flags().BoolVar(&useVPN, "vpn", false, "Enable VPN mode (Gluetun)")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be generated without creating files")
	generateCmd.Flags().BoolVar(&saveProfile, "save-profile", false, "Save configuration as a profile after generation")
	generateCmd.Flags().StringVar(&saveProfileName, "save-as", "", "Profile name when using --save-profile")
}

func runGenerate(t *i18n.I18n) error {
	var loadedProfile *profile.Profile
	var err error

	// Step 0: Load profile if specified
	if profileName != "" {
		fmt.Printf("📋 Loading profile: %s\n", profileName)
		loadedProfile, err = profile.LoadProfile(profileName)
		if err != nil {
			return fmt.Errorf("failed to load profile: %w", err)
		}
		fmt.Printf("✅ Profile loaded: %s\n", loadedProfile.Name)
		if loadedProfile.Description != "" {
			fmt.Printf("   %s\n", loadedProfile.Description)
		}
		fmt.Println()
	}

	// Step 1: Initialize service registry
	registry, err := services.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	// Step 2: Determine VPN setting
	vpnEnabled := useVPN
	if loadedProfile != nil {
		vpnEnabled = loadedProfile.VPN.Enabled
		fmt.Printf("🔒 VPN: %v (from profile)\n", vpnEnabled)
	} else if !dryRun && useVPN == false {
		vpnEnabled, err = prompts.AskVPN(t)
		if err != nil {
			return fmt.Errorf("VPN selection failed: %w", err)
		}
	}

	// Step 3: Select services
	var selectedIDs []string
	if loadedProfile != nil && len(loadedProfile.Services) > 0 {
		selectedIDs = loadedProfile.Services
		fmt.Printf("📦 Services: %s (from profile)\n\n", strings.Join(selectedIDs, ", "))
	} else {
		fmt.Println()
		selectedIDs, err = prompts.SelectServices(t, registry, vpnEnabled)
		if err != nil {
			return fmt.Errorf("service selection failed: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("service selection failed: %w", err)
	}

	if len(selectedIDs) == 0 {
		return fmt.Errorf("%s", t.T("errors.no_services_selected"))
	}

	fmt.Printf("\n✅ %d %s\n\n", len(selectedIDs), t.T("messages.services_selected"))

	// Step 4: Configure environment
	var envConfig *generator.EnvConfig
	if loadedProfile != nil && len(loadedProfile.Environment) > 0 {
		// Use environment from profile
		envConfig = &generator.EnvConfig{
			ComposeProjectName: loadedProfile.Environment["COMPOSE_PROJECT_NAME"],
			ARRPath:            loadedProfile.Environment["ARRPATH"],
			Timezone:           loadedProfile.Environment["TZ"],
			PUID:               loadedProfile.Environment["PUID"],
			PGID:               loadedProfile.Environment["PGID"],
			UMASK:              loadedProfile.Environment["UMASK"],
		}
		
		// Apply VPN config if present
		if vpnEnabled && loadedProfile.VPN.Enabled {
			envConfig.VPNConfig = &generator.VPNConfig{
				ServiceProvider:      loadedProfile.VPN.Provider,
				Type:                 "wireguard",
				WireguardPrivateKey:  loadedProfile.VPN.Password,
				WireguardAddresses:   "",
				WireguardPublicKey:   "",
				PortForwarding:       "off",
				DNSAddress:           "1.1.1.1",
			}
		}
		
		fmt.Println("⚙️  Using environment from profile")
	} else {
		envConfig, err = prompts.ConfigureEnvironment(t, vpnEnabled)
		if err != nil {
			return fmt.Errorf("environment configuration failed: %w", err)
		}
	}

	// Step 5: Validate configuration
	fmt.Println("\n🔍 Validating configuration...")
	validationResult := validateConfiguration(t, registry, selectedIDs, envConfig.ARRPath, outputDir, vpnEnabled)
	
	// Show warnings
	if validationResult.HasWarnings() {
		fmt.Println("\n⚠️  Warnings:")
		for _, warning := range validationResult.Warnings {
			fmt.Printf("   • %s\n", warning.Message)
		}
	}

	// Check for errors
	if validationResult.HasErrors() {
		fmt.Println("\n❌ Validation failed:")
		for _, err := range validationResult.Errors {
			fmt.Printf("   • [%s] %s\n", err.Severity, err.Message)
		}
		return fmt.Errorf("configuration validation failed")
	}

	fmt.Println("✅ Configuration validated")

	// Step 6: Confirm generation
	fmt.Println()
	confirmed, err := prompts.ConfirmGeneration(t)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}

	if !confirmed {
		fmt.Println("\n❌", t.T("messages.generation_cancelled"))
		return nil
	}

	// Step 7: Preview if dry-run
	if dryRun {
		return previewGeneration(t, registry, selectedIDs, envConfig, vpnEnabled)
	}

	// Step 8: Generate files
	if err := generateFiles(t, registry, selectedIDs, envConfig, vpnEnabled); err != nil {
		return err
	}

	// Step 9: Save profile if requested
	if saveProfile || saveProfileName != "" {
		return saveGeneratedProfile(t, selectedIDs, envConfig, vpnEnabled)
	}

	return nil
}

// validateConfiguration runs all validators
func validateConfiguration(t *i18n.I18n, registry *services.Registry, serviceIDs []string, basePath, outputDir string, vpnEnabled bool) *validator.ValidationResult {
	config, err := validator.NewConfig(registry, serviceIDs, basePath, outputDir, vpnEnabled)
	if err != nil {
		result := &validator.ValidationResult{Valid: false}
		result.AddError("config", fmt.Sprintf("Failed to create validation config: %v", err), validator.SeverityCritical)
		return result
	}

	return validator.ValidateAll(config)
}

func previewGeneration(t *i18n.I18n, registry *services.Registry, selectedIDs []string, envConfig *generator.EnvConfig, vpnEnabled bool) error {
	fmt.Println("\n" + "═══════════════════════════════════════════════════════")
	fmt.Println("📋 DRY RUN - Preview Mode")
	fmt.Println("═══════════════════════════════════════════════════════")

	// Preview docker-compose.yml
	composeGen := generator.NewComposeGenerator(registry, outputDir)
	
	composePreview, err := composeGen.Preview(selectedIDs, vpnEnabled)
	if err != nil {
		return fmt.Errorf("compose preview failed: %w", err)
	}

	fmt.Println("\n📄 docker-compose.yml:")
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Println(composePreview)
	fmt.Println("───────────────────────────────────────────────────────")

	// Preview .env
	envGen := generator.NewEnvGenerator(outputDir)
	envPreview, err := envGen.Preview(envConfig)
	if err != nil {
		return fmt.Errorf("env preview failed: %w", err)
	}

	fmt.Println("\n📄 .env:")
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Println(envPreview)
	fmt.Println("───────────────────────────────────────────────────────")

	fmt.Println("\n✅ Preview complete! Run without --dry-run to generate files.")
	return nil
}

func generateFiles(t *i18n.I18n, registry *services.Registry, selectedIDs []string, envConfig *generator.EnvConfig, vpnEnabled bool) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("\n" + "═══════════════════════════════════════════════════════")
	fmt.Println("🚀 Generating files...")
	fmt.Println("═══════════════════════════════════════════════════════")

	// Generate docker-compose.yml
	composeGen := generator.NewComposeGenerator(registry, outputDir)
	
	if vpnEnabled {
		fmt.Println("📡 VPN Mode: Services will use Gluetun network")
	} else {
		fmt.Println("🌉 Bridge Mode: Each service on media network")
	}

	if err := composeGen.Generate(selectedIDs, vpnEnabled, true); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}
	composePath := filepath.Join(outputDir, "docker-compose.yml")
	fmt.Printf("✅ Created: %s\n", composePath)

	// Generate .env
	envGen := generator.NewEnvGenerator(outputDir)
	if err := envGen.Generate(envConfig, true); err != nil {
		return fmt.Errorf("failed to generate .env: %w", err)
	}
	envPath := filepath.Join(outputDir, ".env")
	fmt.Printf("✅ Created: %s\n", envPath)

	// Success message
	fmt.Println("\n" + "═══════════════════════════════════════════════════════")
	fmt.Println("🎉", t.T("messages.generation_complete"))
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("\n📂 Output directory: %s\n", outputDir)
	fmt.Println("\n📝 Next steps:")
	fmt.Println("   1. Review the generated files")
	fmt.Println("   2. Adjust environment variables in .env if needed")
	fmt.Printf("   3. Run: cd %s && docker compose up -d\n", outputDir)
	fmt.Println()

	return nil
}

// saveGeneratedProfile saves the current configuration as a profile
func saveGeneratedProfile(t *i18n.I18n, selectedIDs []string, envConfig *generator.EnvConfig, vpnEnabled bool) error {
	var name string
	
	if saveProfileName != "" {
		name = saveProfileName
	} else {
		// Prompt for profile name
		fmt.Print("\n💾 Profile name: ")
		fmt.Scanln(&name)
	}

	if name == "" {
		fmt.Println("⚠️  Profile name is required. Skipping profile save.")
		return nil
	}

	// Check if profile already exists
	if profile.ProfileExists(name) {
		fmt.Printf("⚠️  Profile '%s' already exists. Overwrite? (y/N): ", name)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" && response != "s" && response != "sim" {
			fmt.Println("ℹ️  Profile save cancelled")
			return nil
		}
	}

	// Create profile
	p := profile.NewProfile(name)
	p.Services = selectedIDs
	p.VPN.Enabled = vpnEnabled
	
	if vpnEnabled && envConfig.VPNConfig != nil {
		p.VPN.Provider = envConfig.VPNConfig.ServiceProvider
		p.VPN.Password = envConfig.VPNConfig.WireguardPrivateKey
	}

	// Save environment variables
	p.Environment = map[string]string{
		"COMPOSE_PROJECT_NAME": envConfig.ComposeProjectName,
		"ARRPATH":              envConfig.ARRPath,
		"TZ":                   envConfig.Timezone,
		"PUID":                 envConfig.PUID,
		"PGID":                 envConfig.PGID,
		"UMASK":                envConfig.UMASK,
	}
	
	p.OutputDir = outputDir

	// Prompt for description
	if saveProfileName == "" {
		fmt.Print("📝 Description (optional): ")
		var desc string
		fmt.Scanln(&desc)
		p.Description = desc
	}

	// Save profile
	if err := profile.SaveProfile(p); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Printf("\n✅ %s: %s\n", t.T("profile.saved_successfully"), name)
	fmt.Println("   Use it with: corsarr generate --profile", name)

	return nil
}
