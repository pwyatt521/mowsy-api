package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"mowsy-api/internal/models"
)

type GeocodioService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewGeocodioService() *GeocodioService {
	apiKey := os.Getenv("GEOCODIO_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: GEOCODIO_API_KEY not set, geocoding will be disabled")
	}

	return &GeocodioService{
		apiKey:  apiKey,
		baseURL: "https://api.geocod.io/v1.7",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type GeocodioResponse struct {
	Results []GeocodioResult `json:"results"`
}

type GeocodioResult struct {
	AddressComponents AddressComponents `json:"address_components"`
	FormattedAddress  string            `json:"formatted_address"`
	Location          Location          `json:"location"`
	Accuracy          float64           `json:"accuracy"`
	Fields            Fields            `json:"fields"`
}

type AddressComponents struct {
	Number            string `json:"number"`
	PredirectionalAbbr string `json:"predirectional"`
	Street            string `json:"street"`
	Suffix            string `json:"suffix"`
	City              string `json:"city"`
	State             string `json:"state"`
	Zip               string `json:"zip"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Fields struct {
	School         *SchoolField `json:"school_districts"`
	Congressional  *CongressionalField `json:"congressional_districts"`
	State          *StateField `json:"state_legislative_districts"`
	Timezone       *TimezoneField `json:"timezone"`
}

type SchoolField struct {
	Elementary []SchoolDistrict `json:"elementary"`
	Secondary  []SchoolDistrict `json:"secondary"`
	Unified    []SchoolDistrict `json:"unified"`
}

type SchoolDistrict struct {
	Name string `json:"name"`
	Code string `json:"lea_code"`
}

type CongressionalField struct {
	Name             string `json:"name"`
	DistrictNumber   int    `json:"district_number"`
	CongressSession  int    `json:"congress_session"`
	CongressYear     int    `json:"congress_year"`
}

type StateField struct {
	House  StateDistrict `json:"house"`
	Senate StateDistrict `json:"senate"`
}

type StateDistrict struct {
	Name           string `json:"name"`
	DistrictNumber string `json:"district_number"`
}

type TimezoneField struct {
	Name               string `json:"name"`
	UTCOffset          int    `json:"utc_offset"`
	ObservsDST         bool   `json:"observes_dst"`
	AbbreviationStd    string `json:"abbreviation"`
	AbbreviationDST    string `json:"abbreviation_dst"`
}

func (g *GeocodioService) geocodeAddress(address string) (*GeocodioResult, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("geocodio API key not configured")
	}

	encodedAddress := url.QueryEscape(address)
	requestURL := fmt.Sprintf("%s/geocode?q=%s&api_key=%s&fields=school_districts", 
		g.baseURL, encodedAddress, g.apiKey)

	resp, err := g.client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make geocoding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("geocoding API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var geocodioResponse GeocodioResponse
	if err := json.Unmarshal(body, &geocodioResponse); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if len(geocodioResponse.Results) == 0 {
		return nil, fmt.Errorf("no geocoding results found for address: %s", address)
	}

	return &geocodioResponse.Results[0], nil
}

func (g *GeocodioService) GeocodeUser(user *models.User) error {
	if user.Address == "" {
		return fmt.Errorf("user address is empty")
	}

	fullAddress := fmt.Sprintf("%s, %s, %s %s", user.Address, user.City, user.State, user.ZipCode)
	result, err := g.geocodeAddress(fullAddress)
	if err != nil {
		return fmt.Errorf("failed to geocode user address: %w", err)
	}

	user.Latitude = &result.Location.Lat
	user.Longitude = &result.Location.Lng

	if result.AddressComponents.Zip != "" {
		user.ZipCode = result.AddressComponents.Zip
	}

	if result.Fields.School != nil {
		if len(result.Fields.School.Elementary) > 0 {
			user.ElementarySchoolDistrictName = result.Fields.School.Elementary[0].Name
			user.ElementarySchoolDistrictCode = result.Fields.School.Elementary[0].Code
		}
	}

	return nil
}

func (g *GeocodioService) GeocodeJob(job *models.Job) error {
	if job.Address == "" {
		return fmt.Errorf("job address is empty")
	}

	result, err := g.geocodeAddress(job.Address)
	if err != nil {
		return fmt.Errorf("failed to geocode job address: %w", err)
	}

	job.Latitude = &result.Location.Lat
	job.Longitude = &result.Location.Lng

	if result.AddressComponents.Zip != "" {
		job.ZipCode = result.AddressComponents.Zip
	}

	if result.Fields.School != nil {
		if len(result.Fields.School.Elementary) > 0 {
			job.ElementarySchoolDistrictName = result.Fields.School.Elementary[0].Name
		}
	}

	return nil
}

func (g *GeocodioService) GeocodeEquipment(equipment *models.Equipment) error {
	if equipment.Address == "" {
		return fmt.Errorf("equipment address is empty")
	}

	result, err := g.geocodeAddress(equipment.Address)
	if err != nil {
		return fmt.Errorf("failed to geocode equipment address: %w", err)
	}

	equipment.Latitude = &result.Location.Lat
	equipment.Longitude = &result.Location.Lng

	if result.AddressComponents.Zip != "" {
		equipment.ZipCode = result.AddressComponents.Zip
	}

	if result.Fields.School != nil {
		if len(result.Fields.School.Elementary) > 0 {
			equipment.ElementarySchoolDistrictName = result.Fields.School.Elementary[0].Name
		}
	}

	return nil
}

func (g *GeocodioService) ReverseGeocode(lat, lng float64) (*GeocodioResult, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("geocodio API key not configured")
	}

	requestURL := fmt.Sprintf("%s/reverse?q=%f,%f&api_key=%s&fields=school_districts", 
		g.baseURL, lat, lng, g.apiKey)

	resp, err := g.client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make reverse geocoding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reverse geocoding API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var geocodioResponse GeocodioResponse
	if err := json.Unmarshal(body, &geocodioResponse); err != nil {
		return nil, fmt.Errorf("failed to parse reverse geocoding response: %w", err)
	}

	if len(geocodioResponse.Results) == 0 {
		return nil, fmt.Errorf("no reverse geocoding results found for coordinates: %f, %f", lat, lng)
	}

	return &geocodioResponse.Results[0], nil
}