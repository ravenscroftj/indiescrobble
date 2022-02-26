package scrobble

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ravenscroftj/indiescrobble/config"
	"github.com/ravenscroftj/indiescrobble/models"
	"github.com/ravenscroftj/indiescrobble/services/micropub"
	"gorm.io/gorm"
)

type Scrobbler struct {
	db *gorm.DB
}

func NewScrobbler(db *gorm.DB) *Scrobbler {
	return &Scrobbler{db: db}
}

func (s *Scrobbler) ValidateType(form *url.Values) error {

	scrobbleType := form.Get("type")
	if _, ok := ScrobbleTypeNames[scrobbleType]; !ok {
		return fmt.Errorf("unknown/invalid scrobble type %v", scrobbleType)
	}

	return nil
}

func (s *Scrobbler) GetItemByID(form *url.Values) (ScrobbleMetaRecord, error) {

	if err := s.ValidateType(form); err != nil {
		return nil, err
	}

	searchEngine := NewSearchProvider(form.Get("type"), s.db)

	item, err := searchEngine.SearchProvider.GetItem(form.Get("item"))

	if err != nil {
		return nil, err
	}

	return item, nil

}

func (s *Scrobbler) Search(form *url.Values) ([]ScrobbleMetaRecord, error) {

	if err := s.ValidateType(form); err != nil {
		return nil, err
	}

	searchEngine := NewSearchProvider(form.Get("type"), s.db)

	query := form.Get("q")

	return searchEngine.SearchProvider.Search(query)

}

func (s *Scrobbler) GetSearchEngineNameForType(scrobbleType string) string {
	return NewSearchProvider(scrobbleType, s.db).SearchProvider.GetName()
}

func (s *Scrobbler) BuildMicroPubPayload(post *models.Post) ([]byte, error) {
	postObj := make(map[string]interface{})
	postObj["type"] = []string{"h-entry"}
	postObj["visibility"] = []string{"public"}

	properties := make(map[string]interface{})

	if post.MediaItem.ThumbnailURL.Valid {
		properties["photo"] = []string{post.MediaItem.ThumbnailURL.String}
	}

	if post.Rating.Valid {
		properties["rating"] = []string{post.Rating.String}
	}

	properties["summary"] = []string{s.GenerateSummary(post)}

	// if the user has enabled it, add the citation e.g. read-of/watch-of/listen-of
	if post.WithWatchOf {
		citationProps := make(map[string]interface{})
		citationProps["name"] = []string{post.MediaItem.DisplayName.String}
		citationProps["uid"] = []string{post.MediaItem.MediaID}
		citationProps["url"] = []string{post.MediaItem.CanonicalURL.String}
		citationProps["indiescrobble-media-id"] = []string{ fmt.Sprintf("%v", post.MediaItem.ID) }
	
		citation := make(map[string]interface{})
		citation["type"] = []string{"h-cite"}
		citation["properties"] = citationProps
	
		// use the appropriate citation property e.g. read-of or watch-of
		properties[ScrobbleCitationProperties[post.PostType]] = citation
	}



	if post.Content.Valid {
		properties["content"] = []string{post.Content.String}
	}

	postObj["properties"] = properties

	return json.MarshalIndent(postObj, "", "  ")
}

func (s *Scrobbler) GenerateSummary(post *models.Post) string {

	rateString := ""

	if post.Rating.Valid {
		rateString = fmt.Sprintf(" and gave it %v/5", post.Rating.String)
	}

	return fmt.Sprintf("%v %v %v%v",
		ScrobbleTypeEmojis[post.PostType],
		ScrobbleTypeVerbs[post.PostType],
		post.MediaItem.DisplayName.String,
		rateString)
}

func (s *Scrobbler) Preview(form *url.Values) (*models.Post, error) {

	if err := s.ValidateType(form); err != nil {
		return nil, err
	}

	item := models.MediaItem{}
	result := s.db.Where(&models.MediaItem{MediaID: form.Get("item")}).First(&item)

	if result.Error != nil {
		return nil, result.Error
	}

	post := models.Post{
		MediaItem: item,
		PostType:  form.Get("type"),
		Content:   sql.NullString{String: form.Get("content"), Valid: form.Get("content") != ""},
		Rating:    sql.NullString{String: form.Get("rating"), Valid: form.Get("rating") != "" },
		WithWatchOf: form.Get("with_watch_of") == "1",
		SharePost: form.Get("share_stats") == "1",
	}

	time, err := time.Parse(config.BROWSER_TIME_FORMAT, form.Get("when"))

	if err == nil {
		post.ScrobbledAt = sql.NullTime{Time: time, Valid: true}
	} else {
		log.Printf("Failed to parse time %v because %v", form.Get("when"), err)
	}

	return &post, nil
}

func (s *Scrobbler) Scrobble(form *url.Values, currentUser *models.BaseUser) (*models.Post, error) {

	if err := s.ValidateType(form); err != nil {
		return nil, err
	}

	item := models.MediaItem{}
	result := s.db.Where(&models.MediaItem{MediaID: form.Get("item")}).First(&item)

	if result.Error != nil {
		log.Printf("Error finding media item with ID %v in db: %v\n", form.Get("item"), result.Error)
		return nil, result.Error
	}

	discovery := micropub.MicropubDiscoveryService{}

	post := models.Post{
		MediaItem: item,
		User:      *currentUser.UserRecord,
		PostType:  form.Get("type"),
		Content:   sql.NullString{String: form.Get("content"), Valid: true},
		Rating:    sql.NullString{String: form.Get("rating"), Valid: true},
	}

	time, err := time.Parse(config.BROWSER_TIME_FORMAT, form.Get("when"))

	if err == nil {
		post.ScrobbledAt = sql.NullTime{Time: time, Valid: true}
	} else {
		log.Printf("Failed to parse time %v because %v", form.Get("when"), err)
	}

	postBody, err := s.BuildMicroPubPayload(&post)

	if err != nil {
		return nil, err
	}

	log.Printf("Send post payload to %v\n", currentUser.Me)

	resp, err := discovery.SubmitMicropub(currentUser, postBody)

	if err != nil {
		log.Printf("Error creating user post: %v\n", err)
		return nil, err
	}

	loc, err := resp.Location()

	if err != nil {
		log.Printf("Error getting Location header from user micropub endpoint: %v\n", err)
		return nil, err
	}

	post.URL = loc.String()
	result = s.db.Create(&post)

	if result.Error != nil {
		log.Printf("Error creating post in database: %v\n", result.Error)
		return nil, result.Error
	}

	return &post, nil
}
