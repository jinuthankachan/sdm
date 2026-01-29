proto-gen:
	@export PATH=$(shell go env GOPATH)/bin:$$PATH && \
	protoc --go_out=. --go_opt=paths=source_relative sdmprotos/annotations.proto

build:
	go build -o bin/sdm cmd/sdm/main.go
	go build -o bin/protoc-gen-sdm ./cmd/protoc-gen-sdm
	