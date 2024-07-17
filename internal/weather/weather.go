package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OwmClient struct {
	OwmApiKey string

	Latitude  float64
	Longitude float64
}

func (owm *OwmClient) GetCurrentCondiitons() (OneCallResponse, error) {
	w, err := owm.getOwmData()

	if err != nil {
		return OneCallResponse{}, fmt.Errorf("Could not retrieve owm data: %v", err)
	}

	return w, nil

}

func (owm *OwmClient) getOwmData() (OneCallResponse, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=hourly,minutely,alerts&units=imperial&appid=%s", owm.Latitude, owm.Longitude, owm.OwmApiKey)

	resp, err := http.Get(url)
	if err != nil {
		return OneCallResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OneCallResponse{}, err
	}

	var data OneCallResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return OneCallResponse{}, err
	}

	if len(data.Daily) < 2 {
		return OneCallResponse{}, fmt.Errorf("not enough data for tomorrow's sunrise")
	}

	return data, nil
}

func FToC(tempf float64) float64 {
	return (tempf - 32) * (5.0 / 9)

}

func GetConditionIcon(iconValue string) string {
	WeatherConditions := map[string]string{
		"01d": "☀️",
		"01n": "🎑",
		"02d": "⛅",
		"02n": "☁️",
		"03d": "🌥️",
		"03n": "☁️",
		"04d": "☁️",
		"04n": "☁️",
		"09d": "🌧️",
		"09n": "🌧️",
		"10d": "🌦️",
		"10n": "🌧️",
		"11d": "⛈️",
		"11n": "⛈️",
		"13d": "🌨️",
		"13n": "🌨️",
		"50d": "🌫️",
		"50n": "🌫️",
	}
	if icon, ok := WeatherConditions[iconValue]; ok {
		return icon
	}
	return ""
}

func LunarPhaseValueToEmoji(lpv float64) (string, error) {
	if lpv < 0 || lpv >= 1 {
		return "", fmt.Errorf("Invalid phase value: %v, expected [0,1)", lpv)
	}
	var lunarPhaseIcon string
	switch lunarPhase := lpv; {
	case lunarPhase == 0:
		lunarPhaseIcon = "🌑"
	case lunarPhase < 0.25:
		lunarPhaseIcon = "🌒"
	case lunarPhase == 0.25:
		lunarPhaseIcon = "🌓"
	case lunarPhase < 0.5:
		lunarPhaseIcon = "🌔"
	case lunarPhase == 0.5:
		lunarPhaseIcon = "🌕"
	case lunarPhase < 0.75:
		lunarPhaseIcon = "🌖"
	case lunarPhase == 0.75:
		lunarPhaseIcon = "🌗"
	case lunarPhase < 1:
		lunarPhaseIcon = "🌘"
	}

	return lunarPhaseIcon, nil
}

type OneCallResponse struct {
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Timezone       string  `json:"timezone"`
	TimezoneOffset int     `json:"timezone_offset"`
	Current        struct {
		Dt         int64   `json:"dt"`
		Sunrise    int     `json:"sunrise"`
		Sunset     int     `json:"sunset"`
		Temp       float64 `json:"temp"`
		FeelsLike  float64 `json:"feels_like"`
		Pressure   int     `json:"pressure"`
		Humidity   int     `json:"humidity"`
		DewPoint   float64 `json:"dew_point"`
		Uvi        float64 `json:"uvi"`
		Clouds     int     `json:"clouds"`
		Visibility int     `json:"visibility"`
		WindSpeed  float64 `json:"wind_speed"`
		WindDeg    int     `json:"wind_deg"`
		WindGust   float64 `json:"wind_gust"`
		Weather    []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"current"`
	Minutely []struct {
		Dt            int64   `json:"dt"`
		Precipitation float64 `json:"precipitation"`
	} `json:"minutely"`
	Daily []struct {
		Dt        int64   `json:"dt"`
		Sunrise   int     `json:"sunrise"`
		Sunset    int     `json:"sunset"`
		Moonrise  int64   `json:"moonrise"`
		Moonset   int64   `json:"moonset"`
		MoonPhase float64 `json:"moon_phase"`
		Temp      struct {
			Day   float64 `json:"day"`
			Min   float64 `json:"min"`
			Max   float64 `json:"max"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"temp"`
		FeelsLike struct {
			Day   float64 `json:"day"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"feels_like"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		DewPoint  float64 `json:"dew_point"`
		WindSpeed float64 `json:"wind_speed"`
		WindDeg   int     `json:"wind_deg"`
		WindGust  float64 `json:"wind_gust"`
		Weather   []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Clouds int     `json:"clouds"`
		Pop    float64 `json:"pop"`
		Rain   float64 `json:"rain"`
		Snow   float64 `json:"snow"`
		Uvi    float64 `json:"uvi"`
	} `json:"daily"`
}
