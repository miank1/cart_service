APP_NAME=cart-service
AIR=$(HOME)/go/bin/air

.PHONY: dev run build start test tidy fmt clean docker

# Development (Live Reload)
dev:
	$(AIR)

run:
	go run ./cmd/main.go

build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) ./cmd/main.go

start: build
	./bin/$(APP_NAME)

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

clean:
	rm -rf bin
	rm -rf tmp

docker:
	docker build -t $(APP_NAME) .
