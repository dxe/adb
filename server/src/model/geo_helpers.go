package model

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/dxe/adb/config"
)

type Location struct {
	Lat float64
	Lng float64
}

// struct for geocoding API: https://developers.google.com/maps/documentation/geocoding/overview
type GeocodeResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status string `json:"status"`
}

func geoCodeAddress(streetAddress string, city string, state string) *Location {
	if config.GooglePlacesBackendAPIKey == "" {
		log.Println("GooglePlacesBackendAPIKey not configured.")
		return nil
	}
	full_address := url.QueryEscape(streetAddress + " " + city + " " + state)
	request := "https://maps.googleapis.com/maps/api/geocode/json?address=" + full_address + "&key=" + config.GooglePlacesBackendAPIKey
	resp, err := http.Get(request)
	if err != nil {
		log.Println("Error geocoding activist location", err)
		return nil
	}
	defer resp.Body.Close()
	var geocode_response GeocodeResponse
	json.NewDecoder(resp.Body).Decode(&geocode_response)
	if len(geocode_response.Results) == 0 {
		log.Printf("No geocoding results found for address %v. Not updating Lat and Lng\n", full_address)
		return nil
	} else {
		return &Location{Lat: geocode_response.Results[0].Geometry.Location.Lat, Lng: geocode_response.Results[0].Geometry.Location.Lng}
	}
}
