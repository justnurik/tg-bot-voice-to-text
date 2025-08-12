.PHONY: build clean

BINARY=bin/vtt
PKG=./cmd/vtt

build:
	go build -o $(BINARY) $(PKG)

clean:
	rm -f $(BINARY)
