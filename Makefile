.PHONY: build clean

APP_ROOT ?= $(PWD)

build: bin/credhub-service-broker

bin/credhub-service-broker: clean
	go build -mod=vendor -o ./bin/credhub-service-broker

clean:
	rm -f bin/*
