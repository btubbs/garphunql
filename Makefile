.PHONY: tests viewcoverage check dep ci

GOLIST=$(shell go list ./...)
GOBIN ?= $(GOPATH)/bin

all: tests check

dep: $(GOBIN)/dep
	$(GOBIN)/dep ensure

tests: dep
	go test .

profile.cov:
	go test -coverprofile=$@

viewcoverage: profile.cov 
	go tool cover -html=$<

vet:
	go vet $(GOLIST)

check: 
$(GOBIN)/golangci-lint:
	go get -v -u github.com/golangci/golangci-lint/cmd/golangci-lint

$(GOBIN)/goveralls:
	go get -v -u github.com/mattn/goveralls

$(GOBIN)/dep:
	go get -v -u github.com/golang/dep/cmd/dep

ci: dep profile.cov vet check $(GOBIN)/golangci-lint $(GOBIN)/goveralls
	$(GOBIN)/goveralls -coverprofile=profile.cov -service=travis-ci
