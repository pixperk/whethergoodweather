# Whether good Weather?

A distributed weather advisory system built with gRPC microservices, providing AI-powered weather insights and real-time streaming recommendations.

## Architecture Overview

This project implements a microservices architecture with the following components:

### Core Services

1. **Weather Service** - Retrieves current weather data from Open-Meteo API
2. **Advisor Service** - Provides AI-powered weather analysis using Google Gemini
3. **API Gateway** - Main server coordinating service communication

### Infrastructure

- **Prometheus** - Metrics collection and monitoring
- **Grafana** - Data visualization and dashboards
- **Docker Compose** - Container orchestration
- **gRPC** - Inter-service communication protocol

## Project Structure

```
├── cmd/
│   ├── cli/           # Interactive command-line interface
│   └── server/        # Main gRPC server
├── services/
│   ├── advisor/       # AI weather advisor service
│   └── weather/       # Weather data retrieval service
├── shared/
│   └── proto/         # Protocol buffer definitions
├── grafana/           # Grafana configuration and dashboards
├── docker-compose.yml # Multi-service deployment
└── prometheus.yml     # Metrics collection configuration
```

## Service Details

### Weather Service

The weather service handles meteorological data retrieval:

- **Endpoint**: `/weather.WeatherService/GetCurrentWeather`
- **Data Source**: Open-Meteo API (free, no API key required)
- **Features**:
  - Current weather conditions
  - Temperature, humidity, wind data
  - Geographic coordinate-based lookup
  - Prometheus metrics integration

### Advisor Service

The advisor service provides intelligent weather insights:

- **Endpoints**: 
  - `/advisor.AdvisorService/GetAdvice` - Single response
  - `/advisor.AdvisorService/StreamAdvice` - Streaming response
- **AI Engine**: Google Gemini 1.5 Flash
- **Features**:
  - Multi-city weather analysis
  - Real-time streaming responses
  - Graceful error handling
  - City geocoding and validation

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- Google Gemini API key

## Environment Setup

1. Set your Google Gemini API key:
```bash
export GEMINI_API_KEY=your_api_key_here
```

2. Clone and navigate to the project:
```bash
git clone <repository-url>
cd effinarounf
```

## Running the Application

### Using Docker Compose (Recommended)

Start all services with monitoring:
```bash
docker-compose up -d
```

This launches:
- Weather and Advisor services on port 8080
- Prometheus on port 9090
- Grafana on port 3000

### Manual Development Setup

1. Start the main server:
```bash
go run cmd/server/main.go
```

2. Access services:
- gRPC server: `localhost:8080`
- Health check: `localhost:8080/health`
- Metrics: `localhost:8080/metrics`

## Usage

### Command Line Interface

The CLI provides interactive access to all services:

```bash
# Interactive mode
go run cmd/cli/main.go

# Direct commands
go run cmd/cli/main.go cities                    # List available cities
go run cmd/cli/main.go weather "New York"       # Get weather for one city
go run cmd/cli/main.go advice "London" "Tokyo"  # Get AI advice for cities
go run cmd/cli/main.go stream "Paris" "Sydney"  # Stream AI advice in real-time
```

### Available Cities

The system supports weather data for 15 major cities worldwide:
- New York, Los Angeles, Chicago, San Francisco, Miami
- London, Paris, Berlin, Barcelona
- Tokyo, Singapore, Sydney
- Toronto, Mumbai, Dubai

### API Integration

#### Weather Service

```go
client := weatherpb.NewWeatherServiceClient(conn)
req := &weatherpb.WeatherRequest{
    Latitude:  40.7128,
    Longitude: -74.0060,
}
resp, err := client.GetCurrentWeather(ctx, req)
```

#### Advisor Service

```go
client := advisorpb.NewAdvisorServiceClient(conn)
req := &advisorpb.AdvisorRequest{
    Cities: []*advisorpb.CityData{
        {Location: "New York"},
        {Location: "London"},
    },
}

// Single response
resp, err := client.GetAdvice(ctx, req)

// Streaming response
stream, err := client.StreamAdvice(ctx, req)
for {
    resp, err := stream.Recv()
    if err == io.EOF {
        break
    }
    // Process streaming chunks
}
```

## Monitoring and Observability

### Prometheus Metrics

Key metrics collected:
- `weather_requests_total` - Total weather API requests
- `weather_request_duration_seconds` - Request latency
- `advisor_requests_total` - Total advisor requests
- `advisor_request_duration_seconds` - AI processing time

Access metrics at `http://localhost:8080/metrics`

### Grafana Dashboards

Pre-configured dashboards available at `http://localhost:3000`:
- Service performance metrics
- Request rates and latency
- Error rates and success ratios
- System resource utilization

Default credentials: `admin/admin`

## Development

### Protocol Buffers

Generate Go code from proto definitions:
```bash
protoc --go_out=. --go-grpc_out=. shared/proto/weather/*.proto
protoc --go_out=. --go-grpc_out=. shared/proto/advisor/*.proto
```

### Testing

Run all tests:
```bash
go test ./...
```

### Building

Create production binaries:
```bash
go build -o bin/server cmd/server/main.go
go build -o bin/cli cmd/cli/main.go
```

## Configuration

### Environment Variables

- `GEMINI_API_KEY` - Google Gemini API key (required)
- `PORT` - Server port (default: 8080)
- `WEATHER_API_URL` - Open-Meteo API base URL

### Service Configuration

Services are configured through:
- Protocol buffer definitions in `shared/proto/`
- Prometheus configuration in `prometheus.yml`
- Grafana datasources in `grafana/provisioning/`

## Error Handling

The system implements robust error handling:
- Graceful degradation when services are unavailable
- Retry mechanisms for external API calls
- Detailed error logging with structured formats
- Health checks for service monitoring

## Performance Considerations

- Connection pooling for external APIs
- Streaming responses for real-time data
- Metrics collection with minimal overhead
- Efficient protobuf serialization
- Resource-optimized Docker containers

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update protocol buffers when changing APIs
4. Maintain backwards compatibility
5. Document significant changes

## License

This project is licensed under the MIT License.
