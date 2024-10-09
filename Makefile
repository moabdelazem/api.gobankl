build:
	@go build -o bin/main.go 

run: build
	@./bin/main.go

watch:
	@air

test:
	@go test -v ./...