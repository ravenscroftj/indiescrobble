package scrobble


type  MetaSearchProvider struct{
	ScrobbleType string
	SearchProvider ScrobbleMetaProvider
}

func NewSearchProvider(scrobbleType string) *MetaSearchProvider{
	provider := &MetaSearchProvider{ScrobbleType: scrobbleType}

	if scrobbleType == SCROBBLE_TYPE_MOVIE {
		provider.SearchProvider = NewIMDBProvider()
	}

	return provider
	
}


func (m *MetaSearchProvider) search(query string) {

}