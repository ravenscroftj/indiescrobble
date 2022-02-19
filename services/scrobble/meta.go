package scrobble

import "gorm.io/gorm"


type  MetaSearchProvider struct{
	ScrobbleType string
	SearchProvider ScrobbleMetaProvider
}

func NewSearchProvider(scrobbleType string, db *gorm.DB) *MetaSearchProvider{
	provider := &MetaSearchProvider{ScrobbleType: scrobbleType}

	if scrobbleType == SCROBBLE_TYPE_MOVIE {
		provider.SearchProvider = NewIMDBProvider(db)
	}

	return provider
	
}
