package scrobble

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/ravenscroftj/indiescrobble/config"
	"github.com/ravenscroftj/indiescrobble/models"
	"gorm.io/gorm"
)

type TMDBMetaRecord struct {
	movie *tmdb.MovieDetails
}

func (r *TMDBMetaRecord) GetCanonicalURL() string {
	return r.movie.Homepage
}

func (r *TMDBMetaRecord) GetDisplayName() string {

	myDate, err := time.Parse("2006-01-02", r.movie.ReleaseDate)

	if err != nil {
		fmt.Printf("Error: %v", err)
		return r.movie.Title

	} else {
		return fmt.Sprintf("%v (%v)", r.movie.Title, myDate.Format("2006"))
	}

}

func (r *TMDBMetaRecord) GetID() string {
	return strconv.FormatInt(r.movie.ID, 10)
}

func (r *TMDBMetaRecord) GetThumbnailURL() string {
	return tmdb.GetImageURL(r.movie.PosterPath, tmdb.W500)
}

type TMDBMetaDataProvider struct {
	tmdbClient *tmdb.Client
	db         *gorm.DB
}

func NewTMDBProvider(db *gorm.DB) (*TMDBMetaDataProvider, error) {

	tmdbClient, err := tmdb.Init(config.GetConfig().GetString("sources.tmdb.apiKey"))

	cache := diskcache.New("cache")
	client := http.Client{Transport: httpcache.NewTransport(cache)}
	tmdbClient.SetClientConfig(client)

	if err != nil {
		return nil, err
	}

	return &TMDBMetaDataProvider{tmdbClient: tmdbClient, db: db}, nil
}

func (t *TMDBMetaDataProvider) GetName() string { return "TMDB" }

func tmdbRecordFromMediaItem(mediaItem *models.MediaItem) TMDBMetaRecord {
	movie := &tmdb.MovieDetails{}
	json.Unmarshal([]byte(mediaItem.Data.String), movie)
	return TMDBMetaRecord{movie: movie}
}

func tmdbRecordToMediaItem(record *TMDBMetaRecord) (*models.MediaItem, error) {

	marshalledTitle, err := json.Marshal(record.movie)

	if err != nil {
		return nil, err
	}

	item := models.MediaItem{
		MediaID:      strconv.FormatInt(record.movie.ID, 10),
		ThumbnailURL: sql.NullString{String: tmdb.GetImageURL(record.movie.PosterPath, tmdb.W500), Valid: true},
		CanonicalURL: sql.NullString{String: record.movie.Homepage, Valid: true},
		DisplayName:  sql.NullString{String: record.movie.Title, Valid: true},
		Data:         sql.NullString{String: string(marshalledTitle), Valid: true},
	}

	return &item, nil
}

func (t *TMDBMetaDataProvider) GetItem(id string) (ScrobbleMetaRecord, error) {

	// see if item is in db first
	item := models.MediaItem{}

	result := t.db.Where(&models.MediaItem{MediaID: id}).First(&item)

	if result.Error == nil {
		record := tmdbRecordFromMediaItem(&item)
		return &record, nil
	}

	tmdbId, err := strconv.Atoi(id)

	if err != nil {
		return nil, err
	}

	title, err := t.tmdbClient.GetMovieDetails(tmdbId, nil)

	if err != nil {
		return nil, err
	}

	// cache the title in db and store
	record := TMDBMetaRecord{movie: title}
	mediaItem, err := tmdbRecordToMediaItem(&record)

	result = t.db.Create(mediaItem)

	if result.Error != nil {
		return nil, result.Error
	}

	if err != nil {
		return nil, err
	}

	return &record, nil

}

func (t *TMDBMetaDataProvider) Search(query string) ([]ScrobbleMetaRecord, error) {

	titles, err := t.tmdbClient.GetSearchMovies(query, nil)

	if err != nil {
		return nil, err
	}

	records := make([]ScrobbleMetaRecord, len(titles.Results))

	for i, title := range titles.Results {
		movie := tmdb.MovieDetails{
			Adult:            title.Adult,
			ID:               title.ID,
			VoteCount:        title.VoteCount,
			Video:            title.Video,
			PosterPath:       title.PosterPath,
			BackdropPath:     title.BackdropPath,
			ReleaseDate:      title.ReleaseDate,
			Overview:         title.Overview,
			OriginalTitle:    title.OriginalTitle,
			OriginalLanguage: title.OriginalLanguage,
			Title:            title.Title,
			VoteAverage:      title.VoteAverage,
		}
		records[i] = &TMDBMetaRecord{movie: &movie}
	}

	return records, nil
}
