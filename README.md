# Weather Advisor Service

gRPC service that provides weather data and AI-generated advice using OpenWeatherMap and Google Gemini.

## Setup

1. Copy `.env.example` to `.env` and add your API keys:
   - Get OpenWeatherMap API key: https://openweathermap.org/api
   - Get Gemini API key: https://ai.google.dev/

2. Run with Docker:
   ```bash
   make docker-up
   ```

3. Or run locally:
   ```bash
   make gen
   make build
   make run
   ```

## Services

- **gRPC Server**: `:8080`
- **Metrics**: `:2112/metrics`
- **Prometheus**: `:9090`
- **Grafana**: `:3000` (admin/admin)

## Testing

```bash
make test
```

## API

### Weather Service
```proto
rpc GetCurrentWeather(WeatherRequest) returns (WeatherResponse);
```

### Advisor Service
```proto
rpc GetAdvice(AdvisorRequest) returns (AdvisorResponse);
```

## Monitoring

Grafana dashboard shows:
- Request rates
- Success rates  
- Response times
- Error rates
