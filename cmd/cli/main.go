package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	advisorpb "github.com/pixperk/effinarounf/shared/proto/advisorpb"
	weatherpb "github.com/pixperk/effinarounf/shared/proto/weatherpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	serverAddr = "localhost:8082"

	// Available cities with their coordinates
	availableCities = map[string][2]float64{
		"New York":      {40.7128, -74.0060},
		"London":        {51.5074, -0.1278},
		"Tokyo":         {35.6762, 139.6503},
		"Paris":         {48.8566, 2.3522},
		"Los Angeles":   {34.0522, -118.2437},
		"Chicago":       {41.8781, -87.6298},
		"Sydney":        {-33.8688, 151.2093},
		"Berlin":        {52.5200, 13.4050},
		"Toronto":       {43.6532, -79.3832},
		"Mumbai":        {19.0760, 72.8777},
		"Dubai":         {25.2048, 55.2708},
		"Singapore":     {1.3521, 103.8198},
		"San Francisco": {37.7749, -122.4194},
		"Miami":         {25.7617, -80.1918},
		"Barcelona":     {41.3851, 2.1734},
	}
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "weather-advisor",
		Short: "AI-Powered Weather Advisor",
		Long: `
┌─────────────────────────────────────────────────────────────┐
│                    Weather Advisor CLI                     │
│                                                             │
│       Get AI-powered weather advice for your cities        │
└─────────────────────────────────────────────────────────────┘
		`,
		Run: func(cmd *cobra.Command, args []string) {
			runInteractiveCLI()
		},
	}

	var listCmd = &cobra.Command{
		Use:   "cities",
		Short: "List all available cities",
		Run: func(cmd *cobra.Command, args []string) {
			listCities()
		},
	}

	var weatherCmd = &cobra.Command{
		Use:   "weather [city]",
		Short: "Get weather for a specific city",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			getWeather(args[0])
		},
	}

	var adviceCmd = &cobra.Command{
		Use:   "advice [cities...]",
		Short: "Get AI advice for cities",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			getAdvice(args, false)
		},
	}

	var streamCmd = &cobra.Command{
		Use:   "stream [cities...]",
		Short: "Get streaming AI advice for cities",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			getAdvice(args, true)
		},
	}

	rootCmd.AddCommand(listCmd, weatherCmd, adviceCmd, streamCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runInteractiveCLI() {
	color.HiCyan("Welcome to Weather Advisor")
	fmt.Println()

	for {
		var action string
		prompt := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{
				"Get Weather for a City",
				"Get AI Weather Advice",
				"Stream AI Advice (Real-time)",
				"List Available Cities",
				"Exit",
			},
		}
		survey.AskOne(prompt, &action)

		switch action {
		case "Get Weather for a City":
			selectAndGetWeather()
		case "Get AI Weather Advice":
			selectAndGetAdvice(false)
		case "Stream AI Advice (Real-time)":
			selectAndGetAdvice(true)
		case "List Available Cities":
			listCities()
		case "Exit":
			color.HiGreen("Goodbye!")
			return
		}
		fmt.Println()
	}
}

func selectAndGetWeather() {
	cities := make([]string, 0, len(availableCities))
	for city := range availableCities {
		cities = append(cities, city)
	}

	var selectedCity string
	prompt := &survey.Select{
		Message: "Select a city:",
		Options: cities,
	}
	survey.AskOne(prompt, &selectedCity)

	getWeather(selectedCity)
}

func selectAndGetAdvice(stream bool) {
	cities := make([]string, 0, len(availableCities))
	for city := range availableCities {
		cities = append(cities, city)
	}

	var selectedCities []string
	prompt := &survey.MultiSelect{
		Message: "Select cities (use space to select, enter to confirm):",
		Options: cities,
	}
	survey.AskOne(prompt, &selectedCities)

	if len(selectedCities) == 0 {
		color.Yellow("⚠️  No cities selected!")
		return
	}

	getAdvice(selectedCities, stream)
}

func listCities() {
	color.HiCyan("\nAvailable Cities:")
	color.Cyan(strings.Repeat("─", 50))

	cities := make([]string, 0, len(availableCities))
	for city := range availableCities {
		cities = append(cities, city)
	}

	for i, city := range cities {
		coords := availableCities[city]
		fmt.Printf("%-3d. %-15s (%.4f, %.6f)\n", i+1, city, coords[0], coords[1])
	}

	color.Cyan(strings.Repeat("─", 50))
	color.HiBlue(fmt.Sprintf("Total: %d cities available", len(cities)))
}

func getWeather(cityName string) {
	coords, exists := availableCities[cityName]
	if !exists {
		color.Red("City '%s' not found! Use 'weather-advisor cities' to see available cities.", cityName)
		return
	}

	color.HiYellow("Getting weather for %s...", cityName)

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		color.Red("Connection failed: %v", err)
		return
	}
	defer conn.Close()

	client := weatherpb.NewWeatherServiceClient(conn)
	req := &weatherpb.WeatherRequest{
		Latitude:  coords[0],
		Longitude: coords[1],
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetCurrentWeather(ctx, req)
	if err != nil {
		color.Red("Weather request failed: %v", err)
		return
	}

	// Display weather info
	color.HiGreen("\nWeather Report for %s", cityName)
	color.Green(strings.Repeat("─", 40))
	fmt.Printf("Temperature: %.1f°C (feels like %.1f°C)\n", resp.Temperature, resp.FeelsLike)
	fmt.Printf("Condition: %s\n", resp.Description)
	fmt.Printf("Humidity: %d%%\n", resp.Humidity)
	fmt.Printf("Wind: %.1f m/s at %d°\n", resp.WindSpeed, resp.WindDeg)
	fmt.Printf("Pressure: %d hPa\n", resp.Pressure)
	color.Green(strings.Repeat("─", 40))
}

func getAdvice(cities []string, stream bool) {
	// Validate all cities
	var cityData []*advisorpb.CityData
	for _, city := range cities {
		if _, exists := availableCities[city]; !exists {
			color.Red("❌ City '%s' not found!", city)
			return
		}
		cityData = append(cityData, &advisorpb.CityData{Location: city})
	}

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		color.Red("❌ Connection failed: %v", err)
		return
	}
	defer conn.Close()

	client := advisorpb.NewAdvisorServiceClient(conn)
	req := &advisorpb.AdvisorRequest{Cities: cityData}

	if stream {
		getStreamingAdvice(client, req, cities)
	} else {
		getNormalAdvice(client, req, cities)
	}
}

func getNormalAdvice(client advisorpb.AdvisorServiceClient, req *advisorpb.AdvisorRequest, cities []string) {
	color.HiYellow("🤖 Getting AI advice for: %s", strings.Join(cities, ", "))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.GetAdvice(ctx, req)
	if err != nil {
		color.Red("❌ Advice request failed: %v", err)
		return
	}

	color.HiGreen("\n🎯 AI Weather Advice")
	color.Green(strings.Repeat("═", 60))
	fmt.Println(resp.Advice)
	color.Green(strings.Repeat("═", 60))
}

func getStreamingAdvice(client advisorpb.AdvisorServiceClient, req *advisorpb.AdvisorRequest, cities []string) {
	color.HiYellow("📡 Streaming AI advice for: %s", strings.Join(cities, ", "))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := client.StreamAdvice(ctx, req)
	if err != nil {
		color.Red("❌ Streaming failed: %v", err)
		return
	}

	color.HiGreen("\n🎯 AI Weather Advice (Streaming)")
	color.Green(strings.Repeat("═", 60))

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {

			if strings.Contains(err.Error(), "context canceled") ||
				strings.Contains(err.Error(), "connection is closing") ||
				strings.Contains(err.Error(), "streaming failed") {
				break
			}
			// Only log unexpected errors
			color.Red("❌ Stream error: %v", err)
			return
		}

		if resp.IsComplete {
			break
		}

		fmt.Print(resp.Chunk)
	}

	color.Green("\n" + strings.Repeat("═", 60))
	color.HiGreen("✅ Advice complete!")
}
