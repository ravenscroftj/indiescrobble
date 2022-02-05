package scrobble

import (
	"fmt"
	"net/http"

	"github.com/StalkR/imdb"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
)

type IMDBMetaRecord struct{
	title imdb.Title
}

func (r *IMDBMetaRecord) GetID() string{
	return r.title.ID
}

func (r *IMDBMetaRecord) GetDisplayName() string{
	return fmt.Sprintf("%v (%v)", r.title.Name, r.title.Year)
}


func (r *IMDBMetaRecord) GetCanonicalURL() string{
	return r.title.URL
}

func (r *IMDBMetaRecord) GetThumbnailURL() string{
	return r.title.Poster.ContentURL
}


type IMDBScrobbleMetadataProvider struct {
	client *http.Client
}

func NewIMDBProvider() *IMDBScrobbleMetadataProvider {
	
	cache := diskcache.New("cache")
	client := &http.Client{Transport: httpcache.NewTransport(cache)}
	return &IMDBScrobbleMetadataProvider{client:client}
}


func (i *IMDBScrobbleMetadataProvider) GetName() string { return "IMDB" }


func (i *IMDBScrobbleMetadataProvider) GetItem(id string) (ScrobbleMetaRecord, error)  { 
	
	title, err := imdb.NewTitle(i.client, id)

	if err != nil{
		return nil, err
	}

	return &IMDBMetaRecord{title: *title}, nil

}

func (i *IMDBScrobbleMetadataProvider) Search(query string) ([]ScrobbleMetaRecord, error) {

	titles, err := imdb.SearchTitle(i.client, query)

	if err != nil{
		return nil, err
	}

	records := make([]ScrobbleMetaRecord, len(titles))

	for i, title := range titles {
		records[i] = &IMDBMetaRecord{title: title}
	}

	return records, nil
}