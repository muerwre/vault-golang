APP?=vault
RELEASE?=0.0.1
GOOS?=linux

COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: check
check: prepare_linter
	# golangcli-lint run -v
	# gometalinter --vendor ./...

.PHONY: build
build: clean
	CGO_ENABLED=1 GOOS=${GOOS} go build -o bin/${APP} \
		-ldflags "-X main.version=${RELEASE} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
			cmd/app/main.go
.PHONY: clean
clean:
	@rm -f bin/${APP}

.PHONY: vendor
vendor: prepare_dep
	dep ensure

.PHONY: serve
serve: build
	bin/${APP} serve

.PHONY: dist
dist: build
	mkdir -p dist
	mkdir -p dist/views
	cp bin/${APP} dist
	cp -rf views bin
	cp config.yaml dist

.PHONY: watch
watch: prepare_watcher
	$(shell ${HAS_WATCHER} serve) 

HAS_DEP := $(shell command -v dep;)
HAS_LINTER := $(shell command -v golangci-lint;)
HAS_WATCHER := $(shell command -v ${GOPATH}/bin/watcher;)

.PHONY: prepare_dep
prepare_dep:
ifndef HAS_DEP
	go get -u -v -d github.com/golang/dep/cmd/dep && \
	go install -v github.com/golang/dep/cmd/dep
endif

.PHONY: prepare_linter
prepare_linter:
ifndef HAS_LINTER
	go get -u -v -d github.com/golangci/golangci-lint/cmd/golangci-lint && \
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint
endif

.PHONE: prepare_watcher
prepare_watcher:
ifndef HAS_WATCHER
	go get github.com/canthefason/go-watcher && \
	go install github.com/canthefason/go-watcher/cmd/watcher
endif