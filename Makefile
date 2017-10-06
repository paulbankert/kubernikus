DATE     := $(shell date +%Y%m%d%H%M%S)
VERSION  ?= v$(DATE)
GOOS     ?= $(shell go env | grep GOOS | cut -d'"' -f2)
IMAGES   := $(shell find . -name Dockerfile\* -type f -maxdepth 1 | cut -f3 -d".")
BINARIES := apiserver kubernikus kubernikusctl wormhole

LDFLAGS := -X github.com/sapcc/kubernikus/pkg/version.VERSION=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS) -s -w"

SRCDIRS  := pkg cmd
PACKAGES := $(shell find $(SRCDIRS) -type d)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

BUILD_ARGS = --build-arg VERSION=$(VERSION)

ifneq ($(http_proxy),)
BUILD_ARGS+= --build-arg http_proxy=$(http_proxy) --build-arg https_proxy=$(https_proxy) --build-arg no_proxy=$(no_proxy)
endif

HAS_GLIDE := $(shell command -v glide;)
HAS_SWAGGER := $(shell command -v swagger;)

.PHONY: all clean code-gen client-gen informer-gen lister-gen

all: $(BINARIES:%=bin/$(GOOS)/%)

bin/%: $(GOFILES) Makefile
	GOOS=$(*D) GOARCH=amd64 go build $(GOFLAGS) -v -i -o $(@D)/$(@F) ./cmd/$(@F)

build: $(IMAGES:%=build/%) 
push: $(IMAGES:%=push/%) 

build/%: $(BINARIES:bin/linux/%) 
	docker build $(BUILD_ARGS) -t sapcc/$(@F):$(VERSION) -f Dockerfile.$(@F) .

push/%: $(IMAGES:%=build/%)
	docker push sapcc/$(@F):$(VERSION)   

pkg/api/rest/operations/kubernikus_api.go: swagger.yml
ifndef HAS_SWAGGER
	$(error You need to have go-swagger installed. Run make bootstrap to fix.)
endif
	swagger generate server --name kubernikus --target pkg/api --model-package models \
		--server-package rest --flag-strategy pflag --principal models.Principal --exclude-main

swagger-generate:
	make -B pkg/api/rest/operations/kubernikus_api.go

# --existing-models github.com/sapcc/kubernikus/pkg/api/models seems not to work in our case
swagger-generate-client:
	swagger generate client --name kubernikus --target pkg/client --client-package kubernikus_generated \
	      --principal models.Principal

clean:
	rm -rf bin/*

include code-generate.mk
code-gen: client-gen informer-gen lister-gen

bootstrap:
ifndef HAS_GLIDE
	brew install glide
endif
ifndef HAS_SWAGGER
	brew tap go-swagger/go-swagger
	brew install go-swagger
endif
