package scrobble

import (
	"fmt"
	"net/url"

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
		return fmt.Errorf("Unknown/invalid scrobble type %v", scrobbleType)
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
	
	

	mediaItem := models.MediaItem{}

	post := models.Post{MediaItem: mediaItem, User: *currentUser.UserRecord, PostType: form.Get("type") }

	discovery.SubmitScrobble(currentUser, &post)

	return &post, nil
}