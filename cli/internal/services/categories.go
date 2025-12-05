package services

// ServiceCategory represents the category of a service
type ServiceCategory string

const (
	CategoryDownload  ServiceCategory = "download"
	CategoryIndexer   ServiceCategory = "indexer"
	CategoryMedia     ServiceCategory = "media"
	CategorySubtitles ServiceCategory = "subtitles"
	CategoryStreaming ServiceCategory = "streaming"
	CategoryRequest   ServiceCategory = "request"
	CategoryTranscode ServiceCategory = "transcode"
	CategoryVPN       ServiceCategory = "vpn"
)

// GetCategoryTranslationKey returns the i18n key for a category
func (c ServiceCategory) GetCategoryTranslationKey() string {
	return "categories." + string(c)
}

// AllCategories returns all available categories in order
func AllCategories() []ServiceCategory {
	return []ServiceCategory{
		CategoryDownload,
		CategoryIndexer,
		CategoryMedia,
		CategorySubtitles,
		CategoryStreaming,
		CategoryRequest,
		CategoryTranscode,
		CategoryVPN,
	}
}
