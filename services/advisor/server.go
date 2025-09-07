package advisor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	advisorpb "github.com/pixperk/effinarounf/shared/proto/advisorpb"
	weatherpb "github.com/pixperk/effinarounf/shared/proto/weatherpb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/api/option"
)

var (
	advisorRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "advisor_requests_total",
			Help: "Total advisor requests",
		},
		[]string{"status"},
	)
	advisorDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "advisor_request_duration_seconds",
			Help: "Advisor request duration",
		},
	)
)

type advisorService struct {
	advisorpb.UnimplementedAdvisorServiceServer
	weatherSvc  weatherpb.WeatherServiceServer
	genaiClient *genai.Client
}

type GeocodeResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
		Admin1    string  `json:"admin1"`
	} `json:"results"`
}

func NewAdvisorService(weatherSvc weatherpb.WeatherServiceServer, geminiAPIKey string) (*advisorService, error) {
	ctx := context.Background()
	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	return &advisorService{
		weatherSvc:  weatherSvc,
		genaiClient: genaiClient,
	}, nil
}

func (s *advisorService) Close() {
	if s.genaiClient != nil {
		s.genaiClient.Close()
	}
}

func (s *advisorService) geocodeCity(ctx context.Context, city *advisorpb.CityData, apiKey string) (float64, float64, error) {
	// Simple hardcoded coordinates for testing - replace with proper geocoding later
	cityCoords := map[string][2]float64{
		"New York":    {40.7128, -74.0060},
		"London":      {51.5074, -0.1278},
		"Tokyo":       {35.6762, 139.6503},
		"Paris":       {48.8566, 2.3522},
		"Los Angeles": {34.0522, -118.2437},
		"Chicago":     {41.8781, -87.6298},
		"Sydney":      {-33.8688, 151.2093},
	}

	if coords, exists := cityCoords[city.Location]; exists {
		return coords[0], coords[1], nil
	}

	// If not in hardcoded list, try the API
	encodedQuery := url.QueryEscape(city.Location)
	apiURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", encodedQuery)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return 0, 0, fmt.Errorf("geocoding failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	var geocodeResp GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geocodeResp); err != nil {
		return 0, 0, fmt.Errorf("decode failed: %v", err)
	}

	if len(geocodeResp.Results) == 0 {
		return 0, 0, fmt.Errorf("location not found: %s (try: New York, London, Tokyo, Paris, Los Angeles, Chicago, Sydney)", city.Location)
	}

	return geocodeResp.Results[0].Latitude, geocodeResp.Results[0].Longitude, nil
}

func (s *advisorService) GetAdvice(ctx context.Context, req *advisorpb.AdvisorRequest) (*advisorpb.AdvisorResponse, error) {
	timer := prometheus.NewTimer(advisorDuration)
	defer timer.ObserveDuration()

	var weatherData []string
	for _, city := range req.Cities {
		lat, lon, err := s.geocodeCity(ctx, city, "")
		if err != nil {
			advisorRequests.WithLabelValues("error").Inc()
			return nil, fmt.Errorf("geocoding failed for %s: %v", city.Location, err)
		}

		weatherReq := &weatherpb.WeatherRequest{Latitude: lat, Longitude: lon}
		weatherResp, err := s.weatherSvc.GetCurrentWeather(ctx, weatherReq)
		if err != nil {
			advisorRequests.WithLabelValues("error").Inc()
			return nil, fmt.Errorf("weather request failed for %s: %v", city.Location, err)
		}

		weatherInfo := fmt.Sprintf("City: %s, Temp: %.1f°C, Condition: %s, Humidity: %d%%, Wind: %.1f m/s",
			weatherResp.Location, weatherResp.Temperature, weatherResp.Description, weatherResp.Humidity, weatherResp.WindSpeed)
		weatherData = append(weatherData, weatherInfo)
	}

	advice, err := s.generateAdvice(ctx, weatherData)
	if err != nil {
		advisorRequests.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("advice generation failed: %v", err)
	}

	advisorRequests.WithLabelValues("success").Inc()
	return &advisorpb.AdvisorResponse{Advice: advice}, nil
}

func (s *advisorService) StreamAdvice(req *advisorpb.AdvisorRequest, stream advisorpb.AdvisorService_StreamAdviceServer) error {
	timer := prometheus.NewTimer(advisorDuration)
	defer timer.ObserveDuration()

	var weatherData []string
	for _, city := range req.Cities {
		lat, lon, err := s.geocodeCity(stream.Context(), city, "")
		if err != nil {
			advisorRequests.WithLabelValues("error").Inc()
			return fmt.Errorf("geocoding failed for %s: %v", city.Location, err)
		}

		weatherReq := &weatherpb.WeatherRequest{Latitude: lat, Longitude: lon}
		weatherResp, err := s.weatherSvc.GetCurrentWeather(stream.Context(), weatherReq)
		if err != nil {
			advisorRequests.WithLabelValues("error").Inc()
			return fmt.Errorf("weather request failed for %s: %v", city.Location, err)
		}

		weatherInfo := fmt.Sprintf("City: %s, Temp: %.1f°C, Condition: %s, Humidity: %d%%, Wind: %.1f m/s",
			weatherResp.Location, weatherResp.Temperature, weatherResp.Description, weatherResp.Humidity, weatherResp.WindSpeed)
		weatherData = append(weatherData, weatherInfo)
	}

	// Stream the advice generation
	err := s.streamAdviceGeneration(stream.Context(), weatherData, stream)
	if err != nil {
		advisorRequests.WithLabelValues("error").Inc()
		return fmt.Errorf("advice generation failed: %v", err)
	}

	advisorRequests.WithLabelValues("success").Inc()
	return nil
}

func (s *advisorService) generateAdvice(ctx context.Context, weatherData []string) (string, error) {
	model := s.genaiClient.GenerativeModel("gemini-1.5-flash")
	prompt := fmt.Sprintf(`Weather advisor. Based on this data provide practical advice:

%s

Include: summary, clothing advice, activity suggestions, warnings. Keep it concise.`, strings.Join(weatherData, "\n"))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Gemini API failed: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	var advice strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			advice.WriteString(string(text))
		}
	}

	return advice.String(), nil
}

func (s *advisorService) streamAdviceGeneration(ctx context.Context, weatherData []string, stream advisorpb.AdvisorService_StreamAdviceServer) error {
	model := s.genaiClient.GenerativeModel("gemini-1.5-flash")
	prompt := fmt.Sprintf(`Weather advisor. Based on this data provide practical advice:

%s

Include: summary, clothing advice, activity suggestions, warnings. Keep it concise.`, strings.Join(weatherData, "\n"))

	// Use streaming generation
	iter := model.GenerateContentStream(ctx, genai.Text(prompt))

	for {
		resp, err := iter.Next()
		if err != nil {
			// Check if it's end of stream
			if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "iterator stopped") {
				// Send completion signal
				return stream.Send(&advisorpb.StreamAdviceResponse{
					Chunk:      "",
					IsComplete: true,
				})
			}
			return fmt.Errorf("streaming failed: %v", err)
		}

		// Extract text from response
		for _, cand := range resp.Candidates {
			for _, part := range cand.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					// Send the text chunk
					err := stream.Send(&advisorpb.StreamAdviceResponse{
						Chunk:      string(text),
						IsComplete: false,
					})
					if err != nil {
						return fmt.Errorf("failed to send chunk: %v", err)
					}
				}
			}
		}
	}
}
