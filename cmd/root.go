package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/woliveiras/corsarr/internal/i18n"
)

var (
	translator *i18n.I18n
	language   string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "corsarr",
	Short: "🏴‍☠️ Corsarr - Navigate the high seas of media automation",
	Long: `Corsarr is a CLI tool to easily configure and deploy your *arr stack
(Radarr, Sonarr, Prowlarr, etc.) with Docker Compose.

Select the services you want, configure your environment,
and Corsarr will generate the docker-compose.yml and .env files for you.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize i18n if not already done
		if translator == nil {
			var err error
			
			// If language not set, prompt user to select
			if language == "" {
				language, err = i18n.SelectLanguage()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error selecting language: %v\n", err)
					language = "en" // Fallback to English
				}
			}

			translator, err = i18n.New(language)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing translator: %v\n", err)
				os.Exit(1)
			}

			// Print welcome message
			fmt.Println(translator.T("messages.welcome"))
			fmt.Println()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&language, "language", "l", "", "Language (en, pt-br, es)")
}

// GetTranslator returns the current translator instance
func GetTranslator() *i18n.I18n {
	return translator
}
