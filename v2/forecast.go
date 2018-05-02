package forecast

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// URL example:  "https://api.forecast.io/forecast/APIKEY/LATITUDE,LONGITUDE,TIME?units=ca"
const (
	BASEURL = "https://api.forecast.io/forecast"
)

type Flags struct {
	DarkSkyUnavailable string   `json:"darksky-unavailable"`
	DarkSkyStations    []string `json:"darksky-stations"`
	DataPointStations  []string `json:"datapoint-stations"`
	ISDStations        []string `json:"isds-stations"`
	LAMPStations       []string `json:"lamp-stations"`
	METARStations      []string `json:"metars-stations"`
	METNOLicense       string   `json:"metnol-license"`
	Sources            []string `json:"sources"`
	Units              string   `json:"units"`
}

type DataPoint struct {
	Time                        float64 `json:"time"`
	Summary                     string  `json:"summary"`
	Icon                        string  `json:"icon"`
	SunriseTime                 float64 `json:"sunrise_time"`
	SunsetTime                  float64 `json:"sunset_time"`
	MoonPhase                   float64 `json:"moon_phase"`
	PrecipIntensity             float64 `json:"precipIntensity"`
	PrecipIntensityMax          float64 `json:"precipIntensityMax"`
	PrecipIntensityMaxTime      float64 `json:"precipIntensityMaxTime"`
	PrecipProbability           float64 `json:"precipProbability"`
	PrecipType                  string  `json:"precipType"`
	PrecipAccumulation          float64 `json:"precipAccumulation"`
	TemperatureLow              float64 `json:"temperatureLow"`
	Temperature                 float64 `json:"temperature"`
	ApparentTemperature         float64 `json:"apparentTemperature"`
	TemperatureLowTime          float64 `json:"temperatureLowTime"`
	TemperatureHigh             float64 `json:"temperatureHigh"`
	TemperatureHighTime         float64 `json:"temperatureHighTime"`
	ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh"`
	ApparentTemperatureHighTime float64 `json:"apparentTemperatureHighTime"`
	ApparentTemperatureLow      float64 `json:"apparentTemperatureLow"`
	ApparentTemperatureLowTime  float64 `json:"apparentTemperatureLowTime"`
	TemperatureMin              float64 `json:"temperature_min"`
	TemperatureMinTime          float64 `json:"temperature_min_time"`
	TemperatureMax              float64 `json:"temperature_max"`
	TemperatureMaxTime          float64 `json:"temperature_max_time"`
	ApparentTemperatureMin      float64 `json:"apparentTemperatureMin"`
	ApparentTemperatureMinTime  float64 `json:"apparentTemperatureMinTime"`
	ApparentTemperatureMax      float64 `json:"apparentTemperatureMax"`
	ApparentTemperatureMaxTime  float64 `json:"apparentTemperatureMaxTime"`
	DewPoint                    float64 `json:"dew_point"`
	Humidity                    float64 `json:"humidity"`
	Pressure                    float64 `json:"pressure"`
	WindSpeed                   float64 `json:"windSpeed"`
	WindGust                    float64 `json:"windGust"`
	WindGustTime                float64 `json:"windGustTime"`
	WindBearing                 float64 `json:"windBearing"`
	CloudCover                  float64 `json:"cloudCover"`
	UVIndex                     int     `json:"uvIndex"`
	UVIndexTime                 int     `json:"uvIndexTime"`
	Ozone                       float64 `json:"ozone"`
	Visibility                  float64 `json:"visibility"`
}

type DataBlock struct {
	Summary string      `json:"summary"`
	Icon    string      `json:"icon"`
	Data    []DataPoint `json:"data"`
}

type alert struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Time        float64 `json:"time"`
	Expires     float64 `json:"expires"`
	URI         string  `json:"uri"`
}

type Forecast struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
	Offset    float64   `json:"offset"`
	Currently DataPoint `json:"currently"`
	Minutely  DataBlock `json:"minutely"`
	Hourly    DataBlock `json:"hourly"`
	Daily     DataBlock `json:"daily"`
	Alerts    []alert   `json:"alerts"`
	Flags     Flags     `json:"flags"`
	APICalls  int       `json:"apicalls"`
	Code      int       `json:"code"`
}

type Units string

const (
	CA   Units = "ca"
	SI   Units = "si"
	US   Units = "us"
	UK   Units = "uk"
	AUTO Units = "auto"
)

func Get(key string, lat string, long string, time string, units Units) (*Forecast, error) {
	res, err := GetResponse(key, lat, long, time, units)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	f, err := FromJSON(body)
	if err != nil {
		return nil, err
	}

	calls, _ := strconv.Atoi(res.Header.Get("X-Forecast-API-Calls"))
	f.APICalls = calls

	return f, nil
}

func FromJSON(jsonBlob []byte) (*Forecast, error) {
	var f Forecast
	err := json.Unmarshal(jsonBlob, &f)
	if err != nil {
		return nil, err
	}
	log.Println(string(jsonBlob))
	for _, v := range f.Daily.Data {
		log.Println("FROM JSON", v)
	}

	return &f, nil
}

// DataBlockType is useful if you want to exclude certain pieces of data from the response
type DataBlockType string

const (
	Currently DataBlockType = "Currently"
	Minutely  DataBlockType = "Minutely"
	Hourly    DataBlockType = "Hourly"
	Daily     DataBlockType = "Daily"
	Alerts    DataBlockType = "Alerts"
	FlagData  DataBlockType = "Flags"
	AlertData DataBlockType = "Alerts"
)

func GetResponse(key string, lat string, long string, time string, units Units) (*http.Response, error) {
	coord := lat + "," + long
	//TODO(mattwarren1234 12/7/2015) : potentially add 'blocks' as a query param
	//exclude=[blocks]:
	// Exclude some number of data blocks from the API response.
	//  This is useful for reducing latency and saving cache space.
	//  [blocks] should be a comma-delimeted list (without spaces) of any of the following:
	//  currently, minutely, hourly, daily, alerts, flags.
	//  (Crafting a request with all of the above blocks excluded is exceedingly silly and not recommended.)

	var url string
	if time == "now" {
		url = BASEURL + "/" + key + "/" + coord + "?units=" + string(units)
	} else {
		url = BASEURL + "/" + key + "/" + coord + "," + time + "?units=" + string(units)
	}

	log.Println(url)

	// if len(exclude) > 0 {
	// 	url = url + "&exclude="
	// 	for i, v := range exclude {
	// 		if i != 0 {
	// 			url = url + ","
	// 		}
	// 		url = url + v
	// 	}
	// }

	res, err := http.Get(url)
	if err != nil {
		return res, err
	}

	return res, nil
}
