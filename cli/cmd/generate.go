package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/woliveiras/corsarr/internal/generator"
	"github.com/woliveiras/corsarr/internal/i18n"
	"github.com/woliveiras/corsarr/internal/prompts"
	"github.com/woliveiras/corsarr/internal/services"
)

var (
	profileName     string
	outputDir       string
	noInteractive   bool
	useVPN          bool
	dryRun          bool
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
}

func runGenerate(t *i18n.I18n) error {
	// TODO: Handle profile loading if profileName is set
	// TODO: Handle non-interactive mode if noInteractive is true
	
	// Step 1: Initialize service registry
	registry, err := services.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	// Step 2: Ask if user wants VPN (unless --vpn flag was used)
	vpnEnabled := useVPN
	if !dryRun && useVPN == false { // Only prompt if not set by flag
		var err error
		vpnEnabled, err = prompts.AskVPN(t)
		if err != nil {
			return fmt.Errorf("VPN selection failed: %w", err)
		}
	}

	// Step 3: Select services
	fmt.Println()
	selectedIDs, err := prompts.SelectServices(t, registry, vpnEnabled)
	if err != nil {
		return fmt.Errorf("service selection failed: %w", err)
	}

	if len(selectedIDs) == 0 {
		return fmt.Errorf(t.T("errors.no_services_selected"))
	}

	fmt.Printf("\n✅ %d %s\n\n", len(selectedIDs), t.T("messages.services_selected"))

	// Step 4: Configure environment
	envConfig, err := prompts.ConfigureEnvironment(t, vpnEnabled)
	if err != nil {
		return fmt.Errorf("environment configuration failed: %w", err)
	}

	// Step 5: Confirm generation
	fmt.Println()
	confirmed, err := prompts.ConfirmGeneration(t)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}

	if !confirmed {
		fmt.Println("\n❌", t.T("messages.generation_cancelled"))
		return nil
	}

	// Step 6: Preview if dry-run
	if dryRun {
		return previewGeneration(t, registry, selectedIDs, envConfig, vpnEnabled)
	}

	// Step 7: Generate files
	return generateFiles(t, registry, selectedIDs, envConfig, vpnEnabled)
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
