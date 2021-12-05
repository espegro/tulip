GOCMD=go
GOBUILD=$(GOCMD) build -ldflags="-s -w"
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

all:clean tulip

tulip:
	$(GOBUILD) tulip.go bloom.go handlers.go file.go

clean:
	$(GOCLEAN)

