package services

import (
	"testing"
)

func TestServiceCategory_GetCategoryTranslationKey(t *testing.T) {
	tests := []struct {
		category ServiceCategory
		expected string
	}{
		{CategoryDownload, "categories.download"},
		{CategoryIndexer, "categories.indexer"},
		{CategoryMedia, "categories.media"},
		{CategoryVPN, "categories.vpn"},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			result := tt.category.GetCategoryTranslationKey()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestAllCategories(t *testing.T) {
	categories := AllCategories()
	
	if len(categories) == 0 {
		t.Fatal("Expected categories, got 0")
	}

	expectedCount := 8 // download, indexer, media, subtitles, streaming, request, transcode, vpn
	if len(categories) != expectedCount {
		t.Errorf("Expected %d categories, got %d", expectedCount, len(categories))
	}

	// Check that all expected categories are present
	expectedCategories := map[ServiceCategory]bool{
		CategoryDownload:  false,
		CategoryIndexer:   false,
		CategoryMedia:     false,
		CategorySubtitles: false,
		CategoryStreaming: false,
		CategoryRequest:   false,
		CategoryTranscode: false,
		CategoryVPN:       false,
	}

	for _, cat := range categories {
		expectedCategories[cat] = true
	}

	for cat, found := range expectedCategories {
		if !found {
			t.Errorf("Category %s not found in AllCategories()", cat)
		}
	}
}
