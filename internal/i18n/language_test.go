package i18n

import (
	"reflect"
	"sort"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSupportedLanguagesCoverDesktopLocales(t *testing.T) {
	want := []string{"en", "es", "pt-BR", "it"}
	got := make([]string, 0, len(SupportedLanguages))
	for _, supported := range SupportedLanguages {
		got = append(got, supported.Code)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected supported languages %v, got %v", want, got)
	}
}

func TestBackendLocaleCatalogsHaveIdenticalMessageIDs(t *testing.T) {
	var baseline []string
	for _, locale := range []string{"en", "es", "pt-br", "it"} {
		data, err := LocaleFS.ReadFile("locales/" + locale + ".yaml")
		if err != nil {
			t.Fatalf("read %s locale: %v", locale, err)
		}
		messages := map[string]any{}
		if err := yaml.Unmarshal(data, &messages); err != nil {
			t.Fatalf("parse %s locale: %v", locale, err)
		}
		keys := localeMessagePaths(messages, "")
		sort.Strings(keys)
		if baseline == nil {
			baseline = keys
			continue
		}
		if !reflect.DeepEqual(keys, baseline) {
			t.Fatalf("locale %s message IDs differ from en\nwant: %v\ngot:  %v", locale, baseline, keys)
		}
	}
}

func TestItalianBackendLocaleCanBeLoaded(t *testing.T) {
	translator, err := New("it")
	if err != nil {
		t.Fatalf("load Italian locale: %v", err)
	}
	if got := translator.T("services_qbittorrent_description"); got == "services_qbittorrent_description" {
		t.Fatal("expected Italian translation for services_qbittorrent_description")
	}
}

func localeMessagePaths(messages map[string]any, prefix string) []string {
	paths := make([]string, 0, len(messages))
	for key, value := range messages {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		nested, ok := value.(map[string]any)
		if ok {
			paths = append(paths, localeMessagePaths(nested, path)...)
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

func TestNormalizeLanguageMatchesSupportedSystemVariants(t *testing.T) {
	tests := map[string]string{
		"en": "en", "en-US": "en", "es-MX": "es", "pt-br": "pt-BR",
		"pt-PT": "pt-BR", "it-IT": "it",
	}
	for input, want := range tests {
		got, err := NormalizeLanguage(input)
		if err != nil || got != want {
			t.Fatalf("normalize %q: want %q, got %q, err=%v", input, want, got, err)
		}
	}
	if _, err := NormalizeLanguage("fr-FR"); err == nil {
		t.Fatal("expected unsupported language to be rejected")
	}
}
