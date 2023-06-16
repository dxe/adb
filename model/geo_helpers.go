package model 
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dxe/adb/config"
)

type Location struct {
	Lat float64 
	Lng float64 
} 

func geoCodeAddress(streetAddress string, city string, state string) (*Location){
	full_address := url.QueryEscape(streetAddress + " " + city + " " + state)
	request := "https://maps.googleapis.com/maps/api/geocode/json?address=" + full_address + "&key=" + config.GooglePlacesAPIKey
	resp, err := http.Get(request)
	if err != nil {
		fmt.Println("Error geocoding activist location", err)
	}
	defer resp.Body.Close()
	var geocode_response GeocodeResponse
	fmt.Println("response body: ", resp.Body)
	json.NewDecoder(resp.Body).Decode(&geocode_response)
	fmt.Println("got response")
	if len(geocode_response.Results) == 0 {
		fmt.Printf("No geocoding results found for address %v. Not updating Lat and Lng", full_address)
		return nil
	} else {
		fmt.Println("about to return")
		return &Location{Lat: geocode_response.Results[0].Geometry.Location.Lat, Lng: geocode_response.Results[0].Geometry.Location.Lng}
	}
}
