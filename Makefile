gen-repo-mocks:
	mockgen -source=internal/repository/repository.go \
	 		-destination=internal/repository/mocks/mock.go