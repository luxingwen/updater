ROOT_DIR    = $(shell pwd)

.PHONY: client
client:
	go build -o bin/updater-client cmd/main.go


.PHONY: build
build: client