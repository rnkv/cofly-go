# dev:
# 	go run cmd/main/main.go
# race:
# 	go run -race -tags=debug cmd/main/main.go
test:
	go test ./...
test-v:
	go test ./... -v