GIT_COMMIT=$(shell git describe --always)

.PHONY: all build clean test

default: build

all: build test

build:
	go build -ldflags "-X github.com/lifei6671/esdunp/main.BuildVersion=${GIT_COMMIT}"

clean:
	rm ./esdump

test:
	go test ./...