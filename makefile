gen:
	protoc --go_out=. --go-grpc_out=. shared/proto/weather.proto
	protoc --go_out=. --go-grpc_out=. shared/proto/advisor.proto

build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client
	go build -o bin/stream-client ./cmd/stream-client
	go build -o bin/weather-advisor ./cmd/cli

run:
	./bin/server

test:
	./bin/client

stream-test:
	./bin/stream-client

cli:
	./bin/weather-advisor

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/
	docker-compose down -v