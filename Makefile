SHELL := /bin/bash

init:
	@echo Initializing the application dependencies
	docker-compose up -d

unit-tests:
	@echo Executing unit tests
	go test ./... -tags=unit

integration-tests:
	@echo Executing integration tests
	go test ./... -tags=integration

acceptance-tests: export SOCKET_ADDR=localhost:5000
acceptance-tests: export MONGO_URI=mongodb://localhost:27017
acceptance-tests: export MONGO_DATABASE=sku_test
acceptance-tests: export TIMEOUT_IN_SECS=2
acceptance-tests: export LOG_FILE_NAME=server_report_file_test.txt
acceptance-tests: export MAX_CONCURRENT_CONNECTIONS=5
acceptance-tests:
	@echo Executing acceptance tests
	go test ./... -tags=acceptance

server-run: export SOCKET_ADDR=localhost:4000
server-run: export MONGO_URI=mongodb://localhost:27017
server-run: export MONGO_DATABASE=sku
server-run: export TIMEOUT_IN_SECS=15
server-run: export LOG_FILE_NAME=server_report_file.txt
server-run: export MAX_CONCURRENT_CONNECTIONS=5
server-run:
	go run cmd/socket-server/main.go
