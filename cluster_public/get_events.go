package cluster_public

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	userAgent = "cluster fetcher"

	recommendedEventUrl   = "https://api.cluster.mu/v1/events/recommended"
	featuredEventUrl      = "https://api.cluster.mu/v1/events/featured"
	inPreparationEventUrl = "https://api.cluster.mu/v1/events/in_preparation"
)

type Event struct {
	Summary Summary `json:"summary"`
}

type RecomendedEventResponse struct {
	Events []RecomendedEvent `json:"events"`
}

type FeaturedEventResponse struct {
	Events []Event `json:"events"`
}

type InPreparationEventResponse struct {
	Events []Event `json:"events"`
	Paging Paging
}

type Paging struct {
	NextToken string `json:"nextToken"`
}

type RecomendedEvent struct {
	PlayerPhotoUrls   []string          `json:"playerPhotoUrls"`
	RequestUserStatus RequestUserStatus `json:"requestUserStatus"`
	Summary           Summary           `json:"summary"`
}

type RequestUserStatus struct {
	IsWatched bool `json:"isWatched"`
}

type Summary struct {
	EventStatus  string      `json:"eventStatus"`
	ID           string      `json:"id"`
	IsTicketing  bool        `json:"isTicketing"`
	Name         string      `json:"name"`
	Owner        Owner       `json:"owner"`
	Reservation  Reservation `json:"reservation"`
	ThumbnailUrl string      `json:"thumbnailUrl"`
	WatchCount   int         `json:"watchCount"`
}

type Owner struct {
	Bio         string `json:"bio"`
	DisplayName string `json:"displayName"`
	IsCertified bool   `json:"isCertified"`
	IsDeleted   bool   `json:"isDeleted"`
	PhotoUrl    string `json:"photoUrl"`
	Rank        string `json:"rank"`
	ShareUrl    string `json:"shareUrl"`
	UserID      string `json:"userId"`
	Username    string `json:"username"`
}

type Reservation struct {
	CloseDatetime time.Time `json:"closeDatetime"`
	OpenDatetime  time.Time `json:"openDatetime"`
}

func newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("X-Cluster-App-Version", "100")
	req.Header.Add("X-Cluster-Build-Version", "1000")
	req.Header.Add("X-Cluster-Device", "Web")
	req.Header.Add("X-Cluster-Platform", "Web")

	return req, nil
}

func doRequest(url string, v interface{}) error {
	client := &http.Client{}
	req, err := newRequest(url)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func GetRecommendedEvents() (*RecomendedEventResponse, error) {
	r := &RecomendedEventResponse{}
	if err := doRequest(recommendedEventUrl, r); err != nil {
		return nil, err
	}

	return r, nil
}

func GetFeaturedEvents() (*FeaturedEventResponse, error) {
	r := &FeaturedEventResponse{}
	if err := doRequest(featuredEventUrl, r); err != nil {
		return nil, err
	}

	return r, nil
}

func GetInPreparationEvents() (*InPreparationEventResponse, error) {
	r := &InPreparationEventResponse{}
	if err := doRequest(inPreparationEventUrl, r); err != nil {
		return nil, err
	}

	return r, nil
}
