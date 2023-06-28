.PHONY: all build install

all: build install

build:
	go build -trimpath -ldflags "-s -w" -o edgo ./cmd/edgo

install:
	go install -trimpath -ldflags "-s -w" ./cmd/edgo

.PHONY: build install