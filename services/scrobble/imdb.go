package scrobble

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"git.jamesravey.me/ravenscroftj/indiescrobble/models"
	"github.com/StalkR/imdb"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"gorm.io/gorm"
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
	db *gorm.DB
}

func NewIMDBProvider(db *gorm.DB) *IMDBScrobbleMetadataProvider {
	
	cache := diskcache.New("cache")
	client := &http.Client{Transport: httpcache.NewTransport(cache)}
	return &IMDBScrobbleMetadataProvider{client:client, db:db}
}


func (i *IMDBScrobbleMetadataProvider) GetName() string { return "IMDB" }


func titleFromMediaItem(mediaItem *models.MediaItem) imdb.Title {
	title := imdb.Title{ID: mediaItem.MediaID, }
	return title
}

func imdbRecordFromMediaItem(mediaItem *models.MediaItem) IMDBMetaRecord {
	title := imdb.Title{}
	json.Unmarshal([]byte(mediaItem.Data.String), &title)
	return IMDBMetaRecord{title:title}
}

func imdbRecordToMediaItem(record *IMDBMetaRecord) (*models.MediaItem, error){

	marshalledTitle, err := json.Marshal(record.title)

	if err != nil{
		return nil, err
	}

	item := models.MediaItem{
		MediaID: record.title.ID,
		ThumbnailURL: sql.NullString{String: record.GetThumbnailURL(), Valid:true},
		CanonicalURL: sql.NullString{String: record.GetCanonicalURL(), Valid: true},
		DisplayName: sql.NullString{String: record.GetDisplayName(), Valid: true},
		Data: sql.NullString{String: string(marshalledTitle), Valid: true},
	}

	return &item, nil
}


func (i *IMDBScrobbleMetadataProvider) GetItem(id string) (ScrobbleMetaRecord, error)  { 

	// see if item is in db first
	item := models.MediaItem{}

	result := i.db.Where(&models.MediaItem{MediaID: id}).First(&item)

	if result.Error == nil{
		record := imdbRecordFromMediaItem(&item)
		return &record, nil
	}
	
	title, err := imdb.NewTitle(i.client, id)

	if err != nil{
		return nil, err
	}

	// cache the title in db and store
	record := IMDBMetaRecord{title: *title}
	mediaItem, err := imdbRecordToMediaItem(&record)
	
	result = i.db.Create(mediaItem)

	if result.Error != nil{
		return nil, result.Error
	}

	if err != nil{
		return nil, err
	}

	return &record, nil

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