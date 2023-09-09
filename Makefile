gen-svc-mocks:
	mockgen -source=internal/service/service.go \
	 		-destination=internal/service/mocks/mock.go

gen-repo-mocks:
	mockgen -source=internal/repository/repository.go \
	 		-destination=internal/repository/mocks/mock.go

gen-storage-mocks:
	mockgen -source=internal/storage/storage.go \
    	 	-destination=internal/storage/mocks/mock.go

html-coverage:
	go test -coverprofile=coverage ./...
	go tool cover -func=coverage
	go tool cover -html=coverage
	del coverage