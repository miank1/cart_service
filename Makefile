APP_NAME=cart-service
PORT=8081

run:
	go run cmd/main.go

build:
	go build -o bin/$(APP_NAME) ./cmd/main.go

start: build
	./bin/$(APP_NAME)

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf bin
