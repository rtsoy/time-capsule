run:
	@docker-compose -f docker-compose.yml up

lint:
	@golangci-lint run

integration-test:
	@echo "Starting test containers..."
	@docker-compose -p time-capsule-test -f docker-compose-test.yml up -d > nul
	@echo "Running tests..."
	@go clean -testcache
	-@go test -v ./tests/...
	@echo "Deleting test containers..."
	@docker-compose -p time-capsule-test -f docker-compose-test.yml rm -fsv > nul

swag:
	@swag init -g cmd/main.go

gen-svc-mocks:
	@mockgen -source=internal/service/service.go \
	 		-destination=internal/service/mocks/mock.go

gen-repo-mocks:
	@mockgen -source=internal/repository/repository.go \
	 		-destination=internal/repository/mocks/mock.go

gen-storage-mocks:
	@mockgen -source=internal/storage/storage.go \
    	 	-destination=internal/storage/mocks/mock.go

html-coverage:
	@go test -coverprofile=coverage ./internal/...
	@go tool cover -func=coverage
	@go tool cover -html=coverage
	@del coverage