package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	BASE_API_URL                 = "api.openweathermap.org"
	API_ENDPOINT_CURRENT_WEATHER = "https://" + BASE_API_URL + "/data/2.5/weather"
	API_ENDPOINT_5DAY_FORECAST   = "https://" + BASE_API_URL + "/data/2.5/forecast"
)

var (
	OPEN_WEATHER_MAP_API_KEY = os.Getenv("GIN_GONIC_OWM_API_KEY")

	requestParams = map[string]string{
		"units": "metric",
		"APPID": OPEN_WEATHER_MAP_API_KEY,
	}

	client = &http.Client{
		Timeout: time.Second * 2,
	}
)

type main struct {
	Temp     float32 `json:"temp"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
	TempMin  float32 `json:"temp_min"`
	TempMax  float32 `json:"temp_max"`
}

type weather struct {
	Id          int    `json:"-"`
	Title       string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type wind struct {
	Speed float32 `json:"speed"`
	Deg   float32 `json:"deg"`
	Gust  float32 `json:"gust"`
}

type CurrentWeatherResponse struct {
	Weather  []weather `json:"weather"`
	Main     main      `json:"main"`
	Wind     wind      `json:"wind"`
	Datetime int       `json:"dt"`
}

type WeatherCommand struct {
	CommandInterface
}

func makeAPICall() {

}

func CurrentWeatherFromCity(city string) *CurrentWeatherResponse {

	req, _ := http.NewRequest("GET", API_ENDPOINT_CURRENT_WEATHER, nil)

	query := req.URL.Query()

	requestParams["q"] = city

	for paramName, paramValue := range requestParams {
		query.Add(paramName, paramValue)
	}

	req.URL.RawQuery = query.Encode()

	response, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	body, readErr := ioutil.ReadAll(response.Body)

	if readErr != nil {
		fmt.Println(readErr)
	}

	defer response.Body.Close()

	cwr := &CurrentWeatherResponse{}

	fmt.Println(string(body))

	err = json.Unmarshal(body, cwr)

	if err != nil {
		fmt.Println(err)
	}

	return cwr

}

func FiveDayForecastFromCity(city string) {

}
