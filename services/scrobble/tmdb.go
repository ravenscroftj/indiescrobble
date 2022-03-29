package scrobble

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	tv    *tmdb.TVDetails
	movie *tmdb.MovieDetails
	ep    *tmdb.TVEpisodeDetails
}

func (r *TMDBMetaRecord) GetCanonicalURL() string {
	if r.movie != nil {
		return r.movie.Homepage
	} else {
		return r.tv.Homepage
	}
}

func (r *TMDBMetaRecord) GetEpisodes() []ScrobbleMetaRecord {
	if r.tv != nil {

		results := make([]ScrobbleMetaRecord, len(r.tv.EpisodeGroups.Results))

		for i, ep := range r.tv.EpisodeGroups.Results {

			epID, err := strconv.ParseInt(ep.ID, 10, 64)

			if err != nil {
				panic(err)
			}

			results[i] = &TMDBMetaRecord{

				ep: &tmdb.TVEpisodeDetails{
					ID:       epID,
					Name:     ep.Name,
					Overview: ep.Description,
				},
			}

		}

		return results

	} else {
		return nil
	}

}

func (r *TMDBMetaRecord) GetDisplayName() string {

	if r.movie != nil {
		myDate, err := time.Parse("2006-01-02", r.movie.ReleaseDate)

		if err != nil {
			fmt.Printf("Error: %v", err)
			return r.movie.Title

		} else {
			return fmt.Sprintf("%v (%v)", r.movie.Title, myDate.Format("2006"))
		}
	} else {
		myDate, err := time.Parse("2006-01-02", r.tv.FirstAirDate)

		if err != nil {
			fmt.Printf("Error: %v", err)
			return r.tv.Name

		} else {
			return fmt.Sprintf("%v (%v)", r.tv.Name, myDate.Format("2006"))
		}
	}

}

func (r *TMDBMetaRecord) GetID() string {

	if r.movie != nil {
		return strconv.FormatInt(r.movie.ID, 10)
	} else {
		return strconv.FormatInt(r.tv.ID, 10)
	}

}

func (r *TMDBMetaRecord) GetThumbnailURL() string {
	if r.movie != nil {
		return tmdb.GetImageURL(r.movie.PosterPath, tmdb.W500)
	} else {
		return tmdb.GetImageURL(r.tv.BackdropPath, tmdb.W500)
	}

}

type TMDBMetaDataProvider struct {
	tmdbClient *tmdb.Client
	db         *gorm.DB
	mediaType  string
}

func NewTMDBProvider(db *gorm.DB, mediaType string) (*TMDBMetaDataProvider, error) {

	tmdbClient, err := tmdb.Init(config.GetConfig().GetString("sources.tmdb.apiKey"))

	cache := diskcache.New("cache")
	client := http.Client{Transport: httpcache.NewTransport(cache)}
	tmdbClient.SetClientConfig(client)

	if err != nil {
		return nil, err
	}

	return &TMDBMetaDataProvider{tmdbClient: tmdbClient, db: db, mediaType: mediaType}, nil
}

func (t *TMDBMetaDataProvider) GetName() string { return "TMDB" }

func tmdbRecordFromMediaItem(mediaItem *models.MediaItem, mediaType string) TMDBMetaRecord {
	movie := &tmdb.MovieDetails{}

	if mediaType == SCROBBLE_TYPE_MOVIE {
		json.Unmarshal([]byte(mediaItem.Data.String), movie)
		return TMDBMetaRecord{movie: movie}
	} else if mediaType == SCROBBLE_TYPE_TV {
		json.Unmarshal([]byte(mediaItem.Data.String), movie)
		return TMDBMetaRecord{movie: movie}
	}

	panic(errors.New("attempt to deserialize invalid media type"))

}

func tmdbRecordToMediaItem(record *TMDBMetaRecord, mediaType string) (*models.MediaItem, error) {

	var item *models.MediaItem

	if mediaType == SCROBBLE_TYPE_MOVIE {
		marshalledTitle, err := json.Marshal(record.movie)

		if err != nil {
			return nil, err
		}

		item = &models.MediaItem{
			MediaID:      strconv.FormatInt(record.movie.ID, 10),
			ThumbnailURL: sql.NullString{String: tmdb.GetImageURL(record.movie.PosterPath, tmdb.W500), Valid: true},
			CanonicalURL: sql.NullString{String: record.movie.Homepage, Valid: true},
			DisplayName:  sql.NullString{String: record.movie.Title, Valid: true},
			Data:         sql.NullString{String: string(marshalledTitle), Valid: true},
		}
	} else if mediaType == SCROBBLE_TYPE_TV {
		marshalledTitle, err := json.Marshal(record.tv)

		if err != nil {
			return nil, err
		}

		item = &models.MediaItem{
			MediaID: strconv.FormatInt(record.tv.ID, 10),
			Data:    sql.NullString{String: string(marshalledTitle), Valid: true},
		}
	}

	return item, nil
}

func (t *TMDBMetaDataProvider) GetItem(id string) (ScrobbleMetaRecord, error) {

	// see if item is in db first
	item := models.MediaItem{}

	result := t.db.Where(&models.MediaItem{MediaID: id}).First(&item)

	if result.Error == nil {
		record := tmdbRecordFromMediaItem(&item, t.mediaType)
		return &record, nil
	}

	tmdbId, err := strconv.Atoi(id)

	if err != nil {
		return nil, err
	}

	if t.mediaType == SCROBBLE_TYPE_MOVIE {
		record, err := t.createMovieRecord(tmdbId)

		if err != nil {
			return nil, err
		}

		return record, nil

	} else if t.mediaType == SCROBBLE_TYPE_TV {
		record, err := t.createTVShowRecord(tmdbId)

		if err != nil {
			return nil, err
		}

		return record, nil
	}

	return nil, fmt.Errorf("Invalid scrobble type provided: %v", t.mediaType)

}

func (t *TMDBMetaDataProvider) createTVShowRecord(tmdbId int) (*TMDBMetaRecord, error) {
	title, err := t.tmdbClient.GetTVDetails(tmdbId, nil)

	if err != nil {
		return nil, err
	}

	// cache the title in db and store
	record := TMDBMetaRecord{tv: title}
	mediaItem, err := tmdbRecordToMediaItem(&record, t.mediaType)

	result := t.db.Create(mediaItem)

	if result.Error != nil {
		return nil, result.Error
	}

	return &record, err
}

func (t *TMDBMetaDataProvider) createMovieRecord(tmdbId int) (*TMDBMetaRecord, error) {
	title, err := t.tmdbClient.GetMovieDetails(tmdbId, nil)

	if err != nil {
		return nil, err
	}

	// cache the title in db and store
	record := TMDBMetaRecord{movie: title}
	mediaItem, err := tmdbRecordToMediaItem(&record, t.mediaType)

	result := t.db.Create(mediaItem)

	if result.Error != nil {
		return nil, result.Error
	}

	return &record, err
}

func (t *TMDBMetaDataProvider) Search(query string) ([]ScrobbleMetaRecord, error) {

	if t.mediaType == SCROBBLE_TYPE_MOVIE {
		return t.searchMovies(query)
	} else if t.mediaType == SCROBBLE_TYPE_TV {
		return t.searchTV(query)
	}

	panic(errors.New("Invalid scrobble type for search engine"))
}

func (t *TMDBMetaDataProvider) searchTV(query string) ([]ScrobbleMetaRecord, error) {

	titles, err := t.tmdbClient.GetSearchTVShow(query, nil)

	if err != nil {
		return nil, err
	}

	records := make([]ScrobbleMetaRecord, len(titles.Results))

	for i, title := range titles.Results {
		episode := tmdb.TVDetails{
			ID:               title.ID,
			VoteCount:        title.VoteCount,
			PosterPath:       title.PosterPath,
			BackdropPath:     title.BackdropPath,
			Overview:         title.Overview,
			OriginalLanguage: title.OriginalLanguage,
			VoteAverage:      title.VoteAverage,
			FirstAirDate:     title.FirstAirDate,
			Name:             title.Name,
		}

		records[i] = &TMDBMetaRecord{tv: &episode}
	}

	return records, nil
}

func (t *TMDBMetaDataProvider) searchMovies(query string) ([]ScrobbleMetaRecord, error) {

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
