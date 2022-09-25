package wirepod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/digital-dream-labs/chipper/pkg/logger"
)

type openWeather struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
}

type openWeatherAPIResponse struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []openWeather `json:"weather"`
	Main    struct {
		Temp float64 `json:"temp"`
	}
	Dt float64 `json:"dt"`
}

func KToC(kelvin float64) float64 {
	return kelvin - 273.16
}

func KToF(kelvin float64) float64 {

	return KToC(kelvin)*9/5 + 32
}

//https://api.openweathermap.org/data/2.5/weather?q={city name},{state code},{country code}&appid={API key}

func getOpenWeather(location string, botUnits string) (condition string, is_forecast string, local_datetime string, speakable_location_string string, temperature string, temperature_unit string) {
	weatherAPIEnabled := os.Getenv("OPENWEATHERAPI_ENABLED")
	weatherAPIKey := os.Getenv("OPENWEATHERAPI_KEY")

	condition = "Snow"
	is_forecast = "false"
	local_datetime = "test"              // preferably local time in UTC ISO 8601 format ("2022-06-15 12:21:22.123")
	speakable_location_string = location // preferably the processed location
	temperature = "120"
	temperature_unit = "C"

	var enabled = false
	if weatherAPIEnabled == "true" && weatherAPIKey != "" {
		logger.Logger("OpenWeather API Enabled")
		enabled = true
	}

	if enabled {
		//Add default country
		if(strings.Count(location, ",") < 2){ 
			location += "," + os.Getenv("OPENWEATHERAPI_CC")
		}
		
		params := url.Values{}
		params.Add("appid", weatherAPIKey)
		params.Add("q", location)
		logger.Logger("Location " + location)

		url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?%s", params.Encode())
		logger.Logger(url)

		resp, err := http.Get(url)

		if err != nil {
			logger.Logger("Error getting data")
			return
		}

		defer resp.Body.Close()
		weatherResponse := openWeatherAPIResponse{}

		err = json.NewDecoder(resp.Body).Decode(&weatherResponse)
		logger.Logger(fmt.Sprintf("%v", weatherResponse))

		if err != nil {
			logger.Logger(err)
			return
		}

		condition = weatherResponse.Weather[0].Main

		var weatherAPICladMap weatherAPICladStruct
		jsonFile, _ := os.ReadFile("./weather-map.json")
		json.Unmarshal(jsonFile, &weatherAPICladMap)
		for _, b := range weatherAPICladMap {
			if b.APIValue == condition {
				condition = b.CladType
				logger.Logger("API Value: " + b.APIValue + ", Clad Type: " + b.CladType)
				break
			}
		}

		is_forecast = "false"
		condition = weatherResponse.Weather[0].Main
		if botUnits == "F" {
			temperature = fmt.Sprintf("%.0f", KToF(weatherResponse.Main.Temp))
		} else {
			temperature = fmt.Sprintf("%.0f", KToC(weatherResponse.Main.Temp))
		}
		
		temperature_unit = "F"
		local_datetime = "test" //fmt.Sprintf("%f", weatherResponse.Dt)
		speakable_location_string = weatherResponse.Name
		logger.Logger(fmt.Sprintf("%s %s %s %s %s %s", is_forecast, condition, temperature, temperature_unit, local_datetime, speakable_location_string))
	}
	return condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit
}

func openWeatherParser(speechText string, botLocation string, botUnits string) (string, string, string, string, string, string) {
	var specificLocation bool
	var apiLocation string
	var speechLocation string
	if strings.Contains(speechText, " in ") {
		splitPhrase := strings.SplitAfter(speechText, " in ")
		speechLocation = strings.TrimSpace(splitPhrase[1])
		if len(splitPhrase) == 3 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2])
		} else if len(splitPhrase) == 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		} else if len(splitPhrase) > 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		}
		logger.Logger("Location parsed from speech: " + "`" + speechLocation + "`")
		specificLocation = true
	} else {
		logger.Logger("No location parsed from speech")
		specificLocation = false
	}
	if specificLocation {
		apiLocation = speechLocation
	} else {
		apiLocation = botLocation
	}

	condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := getOpenWeather(apiLocation, botUnits)
	return condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit
}