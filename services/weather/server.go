package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	weatherpb "github.com/pixperk/effinarounf/shared/proto/weatherpb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	weatherRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_requests_total",
			Help: "Total weather requests",
		},
		[]string{"status"},
	)
	weatherDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "weather_request_duration_seconds",
			Help: "Weather request duration",
		},
	)
)

type weatherService struct {
	weatherpb.UnimplementedWeatherServiceServer
}

func NewWeatherService() weatherpb.WeatherServiceServer {
	return &weatherService{}
}

type OpenMeteoResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		Humidity    int32   `json:"relative_humidity_2m"`
		WindSpeed   float64 `json:"wind_speed_10m"`
		WindDir     int32   `json:"wind_direction_10m"`
		WeatherCode int32   `json:"weather_code"`
	} `json:"current"`
	CurrentUnits struct {
		Temperature string `json:"temperature_2m"`
		WindSpeed   string `json:"wind_speed_10m"`
	} `json:"current_units"`
}

func getWeatherDescription(code int32) string {
	descriptions := map[int32]string{
		0: "clear sky", 1: "mainly clear", 2: "partly cloudy", 3: "overcast",
		45: "fog", 48: "depositing rime fog", 51: "light drizzle", 53: "moderate drizzle",
		55: "dense drizzle", 61: "slight rain", 63: "moderate rain", 65: "heavy rain",
		71: "slight snow", 73: "moderate snow", 75: "heavy snow", 80: "rain showers",
		81: "moderate rain showers", 82: "violent rain showers", 95: "thunderstorm",
	}
	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "unknown"
}

func (s *weatherService) GetCurrentWeather(ctx context.Context, req *weatherpb.WeatherRequest) (*weatherpb.WeatherResponse, error) {
	timer := prometheus.NewTimer(weatherDuration)
	defer timer.ObserveDuration()

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,wind_speed_10m,wind_direction_10m,weather_code&timezone=auto",
		req.Latitude, req.Longitude)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		weatherRequests.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		weatherRequests.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("API status: %d", resp.StatusCode)
	}

	var weatherData OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		weatherRequests.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("decode failed: %v", err)
	}

	response := &weatherpb.WeatherResponse{
		Location:    fmt.Sprintf("%.2f,%.2f", req.Latitude, req.Longitude),
		Temperature: weatherData.Current.Temperature,
		FeelsLike:   weatherData.Current.Temperature, // Open-Meteo doesn't provide feels_like in free tier
		TempMin:     weatherData.Current.Temperature, // Using current temp as min/max
		TempMax:     weatherData.Current.Temperature,
		Pressure:    1013, // Default pressure since not available in free tier
		Humidity:    weatherData.Current.Humidity,
		WindSpeed:   weatherData.Current.WindSpeed,
		WindDeg:     weatherData.Current.WindDir,
		Timestamp:   time.Now().Unix(),
		Description: getWeatherDescription(weatherData.Current.WeatherCode),
	}

	weatherRequests.WithLabelValues("success").Inc()
	return response, nil
}
