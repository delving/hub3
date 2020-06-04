.PHONY: package

NAME:=hub3
MAINTAINER:="Sjoerd Siebinga <sjoerd@delving.eu>"
DESCRIPTION:="Hub3: Linked Open Data Platform"
MODULE:=github.com/delving/hub3/hub3ctl

GO ?= go
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'git describe --abbrev=0 --tags')
GOVERSION:=$(shell sh -c 'go version | cut -d " " -f3')


LDFLAGS:=-X $(MODULE).Version=$(VERSION) -X hub3.hub3ctl.BuildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X hub3.hub3ctl.GitHash=`git rev-parse HEAD` -X hub3.hub3ctl.BuildAgent=`git config user.email`

# var print rule
print-%  : ; @echo $* = $($*)

clean:
	rm -rf build report gin-bin result.bin *.coverprofile */*.coverprofile hub3/hub3.db hub3/models/hub3.db dist server/assets/assets_vfsdata.go

clean-harvesting:
	rm -rf *_ids.txt *_records.xml

clean-build:
	@make clean
	mkdir -p build

create-assets:
	@go generate ./...

run:
	@go run '${LDFLAGS}' hub3/hub3ctl/main.go

build:
	@rm -rf build
	@make clean-build
	@make create-assets
	@go build -a -o build/$(NAME) -ldflags=$(LDFLAGS) $(MODULE)

run-dev:
	gin --path . --build hub3/hub3ctl -buildArgs "-i -tags=dev -ldflags '${LDFLAGS}'" run http

test:
	@richgo test  ./hub3/...

benchmark:
	@richgo test --bench=. -benchmem ./hub3/...

ginkgo:
	@ginkgo -r  -skipPackage go_tests

twatch:
	@ginkgo watch -r -skipPackage go_tests

compose-up:
	@docker-compose up

compose-down:
	@docker-compose down

compose-clean:
	@docker-compose down --volumes

goreport:
	@mkdir -p report
	@rm -rf report/*
	@goreporter -p ../hub3 -r report -e vendor,cmd,utils -f html

protobuffer:
	@make pb.api
	@make pb.viewconfig
	@make pb.webresource
	@make pb.domain
	@make pb.scan

pb.scan:
	@protoc --go_out=. hub3/ead/eadpb/scan.proto

pb.domain:
	@protoc --go_out=. ikuzo/domain/domainpb/domain.proto
	@protoc --go_out=. ikuzo/domain/domainpb/index.proto

pb.webresource:
	@protoc --go_out=. hub3/mediamanager/webresource.proto

pb.api:
	@protoc --go_out=. hub3/fragments/api.proto

pb.viewconfig:
	@protoc --go_out=. hub3/fragments/viewconfig.proto

pprof-dev:
	@pprof --http localhost:6060 -seconds 30 http://localhost:3000/debug/pprof/profile

# ikuzo specific make actions 
uncovered-ikuzo:
	richgo test -coverprofile /tmp/c.out ./... ; uncover /tmp/c.out

IKUZOMODULE:=github.com/delving/hub3/ikuzo/ikuzoctl

IKUZOLDFLAGS:=-X $(IKUZOMODULE)/cmd.version=`git describe --abbrev=0 --tags` -X $(IKUZOMODULE)/cmd.buildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X $(IKUZOMODULE)/cmd.gitHash=`git rev-parse HEAD` -X $(IKUZOMODULE)/cmd.buildAgent=`git config user.email`

build-ikuzo:
	go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

build-ikuzo-static:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

pre-commit:
	go mod tidy
	richgo test -cover -race -count=10 ./...
	golangci-lint run

test-ikuzo:
	richgo test -cover ./ikuzo/...
	golangci-lint run
	
test-no-cache:
	richgo test -cover -count=1 ./ikuzo/...
	golangci-lint run ikuzo

lint-full-ikuzo:
	golangci-lint run --enable=godox --enable=gomnd --enable=maligned --enable=prealloc --enable=gochecknoglobals --enable=gochecknoinits  ikuzo

api-console:
	api-console build -t "RAML 1.0" -a docs/ikuzo/raml/api.raml -o static/api-console

run-dev-ikuzo:
	gin --path . --build ikuzo/ikuzoctl -i -buildArgs "-tags=dev -ldflags '${IKUZOLDFLAGS}'" run serve

ikuzo-generate-assets:
	go run ikuzo/internal/assets/generate.go
