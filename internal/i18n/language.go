package i18n

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
)

// Language represents a supported language
type Language struct {
	Name string
	Code string
	Flag string
}

// SupportedLanguages lists all available languages
var SupportedLanguages = []Language{
	{Name: "English", Code: "en", Flag: "🇺🇸"},
	{Name: "Português Brasileiro", Code: "pt-br", Flag: "🇧🇷"},
	{Name: "Español", Code: "es", Flag: "🇪🇸"},
}

// SelectLanguage prompts the user to select a language
func SelectLanguage() (string, error) {
	// Build options with flags
	options := make([]huh.Option[string], len(SupportedLanguages))
	for i, lang := range SupportedLanguages {
		displayName := fmt.Sprintf("%s %s", lang.Flag, lang.Name)
		options[i] = huh.NewOption(displayName, lang.Code)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your language / Selecione seu idioma / Seleccione su idioma:").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	return selected, nil
}

// DetectSystemLanguage attempts to detect the system language
func DetectSystemLanguage() string {
	// Check LANG environment variable
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LANGUAGE")
	}

	// Map common system locales to our supported languages
	if len(lang) >= 2 {
		langCode := lang[:2]
		switch langCode {
		case "pt":
			return "pt-br"
		case "es":
			return "es"
		case "en":
			return "en"
		}
	}

	return "en" // Default to English
}

// GetLanguageByCode returns the Language struct for a given code
func GetLanguageByCode(code string) (Language, error) {
	for _, lang := range SupportedLanguages {
		if lang.Code == code {
			return lang, nil
		}
	}
	return Language{}, fmt.Errorf("unsupported language code: %s", code)
}
