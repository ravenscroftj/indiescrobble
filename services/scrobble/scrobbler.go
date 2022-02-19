package scrobble

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"git.jamesravey.me/ravenscroftj/indiescrobble/config"
	"git.jamesravey.me/ravenscroftj/indiescrobble/models"
	"git.jamesravey.me/ravenscroftj/indiescrobble/services/micropub"
	"gorm.io/gorm"
)

type Scrobbler struct {
	db *gorm.DB
}

func NewScrobbler(db *gorm.DB) *Scrobbler{
	return &Scrobbler{db:db}
}

func (s *Scrobbler) ValidateType(form *url.Values) error {

	scrobbleType := form.Get("type")
	if _, ok := ScrobbleTypeNames[scrobbleType]; !ok{
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

	if err != nil{
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

func (s *Scrobbler) buildMicroPubPayload(post *models.Post) ([]byte, error) {
	postObj := make(map[string]interface{})
	postObj["type"] = []string{"h-entry"}
	postObj["visibility"] = []string{"public"}

	properties := make(map[string]interface{})

	if post.MediaItem.ThumbnailURL.Valid{
		properties["photo"] = []string{post.MediaItem.ThumbnailURL.String}
	}

	if post.Rating.Valid{
		properties["rating"] = []string{post.Rating.String}
	}

	properties["summary"] = fmt.Sprintf("%v %v and gave it %v/5", ScrobbleTypeVerbs[post.PostType], post.MediaItem.DisplayName.String, post.Rating.String)
	

	citationProps := make(map[string]interface{})
	citationProps["name"] = []string{post.MediaItem.DisplayName.String}
	citationProps["uid"] = []string{post.MediaItem.MediaID}
	citationProps["url"] = []string{post.MediaItem.CanonicalURL.String}
	citationProps["indiescrobble-id"] = post.MediaItem.ID

	citation := make(map[string]interface{})
	citation["type"] = []string{"h-cite"}
	citation["properties"] = citationProps

	// use the appropriate citation property e.g. read-of or watch-of
	properties[ScrobbleCitationProperties[post.PostType]] = citation

	if post.Content.Valid{
		properties["content"] = []string{post.Content.String}
	}

	postObj["properties"] = properties


	return json.Marshal(postObj)
}

func (s *Scrobbler) Scrobble(form *url.Values, currentUser *models.BaseUser) (*models.Post, error) {

	if err := s.ValidateType(form); err != nil{
		return nil, err
	}
	
	item := models.MediaItem{}
	result := s.db.Where(&models.MediaItem{MediaID: form.Get("item")}).First(&item)

	if result.Error != nil{
		return nil, result.Error
	}

	discovery := micropub.MicropubDiscoveryService{}
	

	post := models.Post{
		MediaItem: item, 
		User: *currentUser.UserRecord, 
		PostType: form.Get("type"),
		Content: sql.NullString{String: form.Get("content"), Valid: true},
		Rating:  sql.NullString{String: form.Get("rating"), Valid: true},
	}


	time, err := time.Parse(config.BROWSER_TIME_FORMAT, form.Get("when"))

	if err == nil{
		post.ScrobbledAt = sql.NullTime{Time: time, Valid: true}
	}else{
		fmt.Errorf("Failed to parse time %v because %v",form.Get("when"), err )
	}

	postBody, err := s.buildMicroPubPayload(&post)


	fmt.Printf("Post body: %v\n", string(postBody))

	if err != nil{
		return nil, err
	}


	resp, err := discovery.SubmitMicropub(currentUser, postBody)

	if err != nil{
		return nil, err
	}

	loc, err := resp.Location()

	if err != nil{
		return nil, err
	}

	post.URL = loc.String()
	result = s.db.Create(&post)

	if result.Error != nil{
		return nil, result.Error
	}

	return &post, nil
}


