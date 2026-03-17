.PHONY: build test vet run demo clean docker-up docker-down

build:
	go build -o bin/streamlens ./cmd/streamlens

test:
	go test ./...

vet:
	go vet ./...

run: build
	./bin/streamlens serve

demo: build
	./bin/streamlens demo

clean:
	rm -rf bin/

docker-up:
	docker compose -f deployments/docker-compose.yml up --build -d

docker-down:
	docker compose -f deployments/docker-compose.yml down
