package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/pixperk/effinarounf/services/advisor"
	"github.com/pixperk/effinarounf/services/weather"
	advisorpb "github.com/pixperk/effinarounf/shared/proto/advisorpb"
	weatherpb "github.com/pixperk/effinarounf/shared/proto/weatherpb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY not set")
	}

	log.Printf("Loaded GEMINI_API_KEY: %s...", geminiAPIKey[:10])

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Metrics server on :2112")
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Listen failed: %v", err)
	}

	s := grpc.NewServer()

	weatherSvc := weather.NewWeatherService()
	weatherpb.RegisterWeatherServiceServer(s, weatherSvc)

	advisorSvc, err := advisor.NewAdvisorService(weatherSvc, geminiAPIKey)
	if err != nil {
		log.Fatalf("Advisor service failed: %v", err)
	}
	defer advisorSvc.Close()
	advisorpb.RegisterAdvisorServiceServer(s, advisorSvc)

	log.Println("gRPC server on :8080")
	log.Fatal(s.Serve(lis))
}
