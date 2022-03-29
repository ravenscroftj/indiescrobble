package scrobble

import "gorm.io/gorm"

type MetaSearchProvider struct {
	ScrobbleType   string
	SearchProvider ScrobbleMetaProvider
}

func NewSearchProvider(scrobbleType string, db *gorm.DB) (*MetaSearchProvider, error) {
	provider := &MetaSearchProvider{ScrobbleType: scrobbleType}

	// if scrobbleType == SCROBBLE_TYPE_MOVIE {
	// 	provider.SearchProvider = NewIMDBProvider(db)
	// }

	if scrobbleType == SCROBBLE_TYPE_MOVIE || scrobbleType == SCROBBLE_TYPE_TV {
		var err error
		provider.SearchProvider, err = NewTMDBProvider(db, scrobbleType)

		if err != nil {
			return nil, err
		}
	}

	return provider, nil

}
