package i18n

import (
	"fmt"
	"os"
	"strings"

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
	{Name: "Español", Code: "es", Flag: "🇪🇸"},
	{Name: "Português Brasileiro", Code: "pt-BR", Flag: "🇧🇷"},
	{Name: "Italiano", Code: "it", Flag: "🇮🇹"},
}

// NormalizeLanguage maps supported regional variants to the canonical locale
// identifiers persisted by Corsarr Desktop.
func NormalizeLanguage(code string) (string, error) {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(code), "_", "-"))
	switch {
	case normalized == "en" || strings.HasPrefix(normalized, "en-"):
		return "en", nil
	case normalized == "es" || strings.HasPrefix(normalized, "es-"):
		return "es", nil
	case normalized == "pt" || strings.HasPrefix(normalized, "pt-"):
		return "pt-BR", nil
	case normalized == "it" || strings.HasPrefix(normalized, "it-"):
		return "it", nil
	default:
		return "", fmt.Errorf("unsupported language code: %s", code)
	}
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
				Title("Select your language / Selecione seu idioma / Seleccione su idioma / Seleziona la lingua:").
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

	if normalized, err := NormalizeLanguage(lang); err == nil {
		return normalized
	}

	return "en" // Default to English
}

// GetLanguageByCode returns the Language struct for a given code
func GetLanguageByCode(code string) (Language, error) {
	normalized, err := NormalizeLanguage(code)
	if err != nil {
		return Language{}, err
	}
	for _, lang := range SupportedLanguages {
		if lang.Code == normalized {
			return lang, nil
		}
	}
	return Language{}, fmt.Errorf("unsupported language code: %s", code)
}
