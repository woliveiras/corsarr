package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/woliveiras/corsarr/internal/profile"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles",
	Long:  "Save, load, list, delete, export and import configuration profiles",
}

var profileSaveCmd = &cobra.Command{
	Use:   "save [name]",
	Short: "Save the current configuration as a profile",
	Long:  "Save the current configuration as a named profile for later reuse",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t := GetTranslator()
		name := args[0]

		description, _ := cmd.Flags().GetString("description")

		// Check if profile already exists
		if profile.ProfileExists(name) {
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				return fmt.Errorf("%s", t.T("profile.already_exists"))
			}
		}

		// Create a new profile
		// Note: In real usage, this would be populated from interactive prompts or flags
		// For now, we create an example profile
		p := profile.NewProfile(name)
		p.Description = description

		if err := profile.SaveProfile(p); err != nil {
			return fmt.Errorf("%s: %w", t.T("profile.save_failed"), err)
		}

		fmt.Printf("✅ %s: %s\n", t.T("profile.saved_successfully"), name)
		return nil
	},
}

var profileLoadCmd = &cobra.Command{
	Use:   "load [name]",
	Short: "Load a configuration profile",
	Long:  "Load a previously saved configuration profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t := GetTranslator()
		name := args[0]

		p, err := profile.LoadProfile(name)
		if err != nil {
			return fmt.Errorf("%s: %w", t.T("profile.load_failed"), err)
		}

		fmt.Printf("✅ %s: %s\n", t.T("profile.loaded_successfully"), name)
		fmt.Printf("\n📋 %s:\n", t.T("profile.details"))
		fmt.Printf("  %s: %s\n", t.T("profile.name"), p.Name)
		if p.Description != "" {
			fmt.Printf("  %s: %s\n", t.T("profile.description"), p.Description)
		}
		fmt.Printf("  %s: %s\n", t.T("profile.created_at"), p.CreatedAt.Format(time.RFC3339))
		fmt.Printf("  %s: %s\n", t.T("profile.updated_at"), p.UpdatedAt.Format(time.RFC3339))
		fmt.Printf("  %s: %s\n", t.T("profile.version"), p.Version)
		fmt.Printf("  %s: %v\n", t.T("profile.vpn_enabled"), p.VPN.Enabled)
		if len(p.Services) > 0 {
			fmt.Printf("  %s: %s\n", t.T("profile.services"), strings.Join(p.Services, ", "))
		}

		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	Long:  "Display a list of all saved configuration profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := GetTranslator()

		profiles, err := profile.ListProfiles()
		if err != nil {
			return fmt.Errorf("%s: %w", t.T("profile.list_failed"), err)
		}

		if len(profiles) == 0 {
			fmt.Printf("ℹ️  %s\n", t.T("profile.no_profiles"))
			return nil
		}

		fmt.Printf("📋 %s (%d):\n\n", t.T("profile.saved_profiles"), len(profiles))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			t.T("profile.name"),
			t.T("profile.services_count"),
			t.T("profile.updated_at"),
			t.T("profile.description"))
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			strings.Repeat("-", 20),
			strings.Repeat("-", 15),
			strings.Repeat("-", 20),
			strings.Repeat("-", 30))

		for _, p := range profiles {
			updatedAt := p.UpdatedAt.Format("2006-01-02 15:04")
			servicesCount := fmt.Sprintf("%d", len(p.Services))
			description := p.Description
			if len(description) > 30 {
				description = description[:27] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				p.Name, servicesCount, updatedAt, description)
		}

		_ = w.Flush()
		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a configuration profile",
	Long:  "Remove a saved configuration profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t := GetTranslator()
		name := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("⚠️  %s '%s'? (y/N): ", t.T("profile.confirm_delete"), name)
			var response string
			_, _ = fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" && response != "s" && response != "sim" {
				fmt.Printf("ℹ️  %s\n", t.T("profile.delete_cancelled"))
				return nil
			}
		}

		if err := profile.DeleteProfile(name); err != nil {
			return fmt.Errorf("%s: %w", t.T("profile.delete_failed"), err)
		}

		fmt.Printf("✅ %s: %s\n", t.T("profile.deleted_successfully"), name)
		return nil
	},
}

var profileExportCmd = &cobra.Command{
	Use:   "export [name] [output-file]",
	Short: "Export a profile to a file",
	Long:  "Export a configuration profile to a JSON file for sharing or backup",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		t := GetTranslator()
		name := args[0]
		outputPath := args[1]

		// Ensure .json extension
		if filepath.Ext(outputPath) != ".json" {
			outputPath += ".json"
		}

		if err := profile.ExportProfile(name, outputPath); err != nil {
			return fmt.Errorf("%s: %w", t.T("profile.export_failed"), err)
		}

		absPath, _ := filepath.Abs(outputPath)
		fmt.Printf("✅ %s: %s\n", t.T("profile.exported_successfully"), absPath)
		return nil
	},
}

var profileImportCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import a profile from a file",
	Long:  "Import a configuration profile from a JSON or YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t := GetTranslator()
		inputPath := args[0]

		p, err := profile.ImportProfile(inputPath)
		if err != nil {
			return fmt.Errorf("%s: %w", t.T("profile.import_failed"), err)
		}

		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			p.Name = name
		}

		// Check if profile already exists
		if profile.ProfileExists(p.Name) {
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				return fmt.Errorf("%s", t.T("profile.already_exists"))
			}
		}

		if err := profile.SaveProfile(p); err != nil {
			return fmt.Errorf("%s: %w", t.T("profile.save_failed"), err)
		}

		fmt.Printf("✅ %s: %s\n", t.T("profile.imported_successfully"), p.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)

	// Add subcommands
	profileCmd.AddCommand(profileSaveCmd)
	profileCmd.AddCommand(profileLoadCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileExportCmd)
	profileCmd.AddCommand(profileImportCmd)

	// Flags for save command
	profileSaveCmd.Flags().StringP("description", "d", "", "Profile description")
	profileSaveCmd.Flags().BoolP("force", "f", false, "Overwrite existing profile")

	// Flags for delete command
	profileDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Flags for import command
	profileImportCmd.Flags().StringP("name", "n", "", "Override profile name")
	profileImportCmd.Flags().BoolP("force", "f", false, "Overwrite existing profile")
}
