package weather

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	OPEN_WEATHER_MAP_API_KEY     = "2f83657530b7825a2abf5d2f9fa37b60"
	BASE_API_URL                 = "api.openweathermap.org"
	API_ENDPOINT_CURRENT_WEATHER = "https://" + BASE_API_URL + "/data/2.5/weather"
	API_ENDPOINT_5DAY_FORECAST   = "https://" + BASE_API_URL + "/data/2.5/forecast"
)

var (
	requestParams = map[string]string{
		"units": "metric",
		"APPID": OPEN_WEATHER_MAP_API_KEY,
	}

	client = &http.Client{
		Timeout: time.Second * 2,
	}
)

func makeAPICall() {

}

func CurrentWeatherFromCity(city string) {
	fmt.Println("Hello world")

	req, _ := http.NewRequest("GET", API_ENDPOINT_CURRENT_WEATHER, nil)

	query := req.URL.Query()

	for paramName, paramValue := range requestParams {
		query.Add(paramName, paramValue)
	}

	query.Add("q", city)

	req.URL.RawQuery = query.Encode()

	fmt.Println(req.URL.String())

	response, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	body, readErr := ioutil.ReadAll(response.Body)

	fmt.Println(readErr)

	fmt.Println(string(body))

	defer response.Body.Close()

}

func FiveDayForecastFromCity(city string) {

}
