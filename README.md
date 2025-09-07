# Weather Advisor Microservices

A distributed weather advisory system built with gRPC microservices, providing AI-powered weather insights and real-time streaming recommendations.

## Architecture Overview

This project implements a microservices architecture with the following components:

### Core Services

1. **Weather Service** - Retrieves current weather data from Open-Meteo API
2. **Advisor Service** - Provides AI-powered weather analysis using Google Gemini
3. **Main Server** - gRPC server hosting both services with metrics endpoint

### Infrastructure

- **Prometheus** - Metrics collection and monitoring (port 9090)
- **Grafana** - Data visualization and dashboards (port 3000)
- **Docker Compose** - Container orchestration for monitoring stack
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
├── docker-compose.yml # Monitoring stack deployment
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

### Hybrid Setup (Recommended)

Start monitoring stack with Docker Compose:
```bash
docker-compose up -d
```

This launches:
- Prometheus on port 9090
- Grafana on port 3000

Then run the main server locally:
```bash
go run cmd/server/main.go
```

This starts:
- gRPC server on port 8082
- Metrics endpoint on port 2113

### Manual Development Setup

1. Start the main server:
```bash
go run cmd/server/main.go
```

2. Access services:
- gRPC server: `localhost:8082`
- Metrics endpoint: `localhost:2113/metrics`

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
conn, err := grpc.Dial("localhost:8082", grpc.WithInsecure())
client := weatherpb.NewWeatherServiceClient(conn)
req := &weatherpb.WeatherRequest{
    Latitude:  40.7128,
    Longitude: -74.0060,
}
resp, err := client.GetCurrentWeather(ctx, req)
```

#### Advisor Service

```go
conn, err := grpc.Dial("localhost:8082", grpc.WithInsecure())
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

Access metrics at `http://localhost:2113/metrics`

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

### Service Configuration

Services are configured through:
- Protocol buffer definitions in `shared/proto/`
- Prometheus configuration in `prometheus.yml` (scrapes localhost:2113)
- Grafana datasources in `grafana/provisioning/`

### Port Configuration

Current port allocation:
- **gRPC Server**: 8082
- **Metrics Endpoint**: 2113
- **Prometheus**: 9090
- **Grafana**: 3000

## Troubleshooting

### Common Issues

1. **Port conflicts**: If you see "bind: address already in use", check if other services are using ports 8082 or 2113
2. **Metrics not showing**: Ensure the server is running and Prometheus is configured to scrape localhost:2113
3. **Advisor service errors**: Check that GEMINI_API_KEY is set correctly

### Monitoring Health

Check service status:
```bash
# Test gRPC server
go run cmd/cli/main.go cities

# Test metrics endpoint
curl http://localhost:2113/metrics

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets
```

## Error Handling

The system implements robust error handling:
- Graceful degradation when external APIs are unavailable
- Streaming error recovery without stopping the entire stream
- Detailed error logging with structured formats
- Comprehensive metrics for monitoring service health

## Performance Considerations

- Connection pooling for external APIs
- Streaming responses for real-time data
- Metrics collection with minimal overhead
- Efficient protobuf serialization
- Optimized Docker containers for monitoring stack

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update protocol buffers when changing APIs
4. Maintain backwards compatibility
5. Document significant changes

## License

This project is licensed under the MIT License.
