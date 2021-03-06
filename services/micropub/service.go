package micropub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ravenscroftj/indiescrobble/models"
)

const (
	USER_AGENT_STRING = "IndieScrobble (indiescrobble.club)"
)

type MicropubDiscoveryService struct {
}

func (m *MicropubDiscoveryService) doGet(url string) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", USER_AGENT_STRING)

	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (m *MicropubDiscoveryService) doAuthGet(url string, bearerToken string) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", USER_AGENT_STRING)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", bearerToken))

	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (m *MicropubDiscoveryService) findMicropubEndpoint(me string) (string, error) {

	res, err := m.doGet(me)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	// Find the review items
	s := doc.Find("link[rel=micropub]")

	if s.Length() < 1 {
		return "", fmt.Errorf("no micropub endpoint found for %v", me)
	}

	// parse the returned URL
	endpointUrl, err := url.Parse(s.AttrOr("href", ""))

	if err != nil {
		return "", err
	}

	if !endpointUrl.IsAbs() {

		if strings.HasPrefix(endpointUrl.Path, "/") {

			newUrl := *res.Request.URL
			newUrl.Path = endpointUrl.Path

			return newUrl.String(), nil
		} else {
			return res.Request.URL.String() + endpointUrl.Path, nil
		}

	} else {
		return endpointUrl.String(), nil
	}
}

func (m *MicropubDiscoveryService) getMicropubConfig(endpoint string, authToken string) (*MicroPubConfig, error) {

	configUrl, err := url.Parse(endpoint)

	if err != nil {
		return nil, err
	}

	q := configUrl.Query()
	q.Set("q", "config")
	configUrl.RawQuery = q.Encode()

	res, err := m.doAuthGet(configUrl.String(), authToken)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	config := MicroPubConfig{}

	err = json.Unmarshal(body, &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

/* Discover endpoints for given me/domain identifier */
func (m *MicropubDiscoveryService) Discover(me string, authToken string) (*MicroPubConfig, error) {

	endpoint, err := m.findMicropubEndpoint(me)

	if err != nil {
		log.Printf("Failed to get endpoint: %v\n", err)
		return nil, err
	}

	// get endpoint config
	config, err := m.getMicropubConfig(endpoint, authToken)

	if err != nil {
		log.Printf("Failed to get configuration: %v\n", err)
		return nil, err
	}

	return config, nil

}

/* Send micropub to endpoint */
func (m *MicropubDiscoveryService) SubmitMicropub(currentUser *models.BaseUser, payload []byte) (*http.Response, error) {

	endpoint, err := m.findMicropubEndpoint(currentUser.Me)

	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(payload)

	req, err := http.NewRequest("POST", endpoint, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", USER_AGENT_STRING)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", currentUser.Token))
	req.Header.Add("Content-Type", "application/json")

	return http.DefaultClient.Do(req)

}
