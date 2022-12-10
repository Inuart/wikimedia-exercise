build:
	@go build -o shortdescriptionapi cmd/main.go

run-demo:
	@CONTACT_INFO=eduard.castany@gmail.com ADDR=localhost:8080 CACHE_SIZE=10 CACHED_RESULT_TTL=1h go run cmd/main.go

test:
	@go test -race -short

test-integration:
	@go test -race