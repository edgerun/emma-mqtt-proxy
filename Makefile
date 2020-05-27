# Go parameters
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

CURDIR=$(shell pwd)
export GOBIN := $(CURDIR)/bin

all: test build-all

build-all:
	$(GOINSTALL) ./...

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf bin/
