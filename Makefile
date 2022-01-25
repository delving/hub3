.PHONY: package

NAME:=hub3
MAINTAINER:="Sjoerd Siebinga <sjoerd@delving.eu>"
DESCRIPTION:="Hub3: Linked Open Data Platform"
MODULE:=github.com/delving/hub3/ikuzoctl

GO ?= go
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'git describe --abbrev=0 --tags')
GOVERSION:=$(shell sh -c 'go version | cut -d " " -f3')

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

benchmark:
	@richgo test --bench=. -benchmem ./...

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
	@make pb.domain
	@make pb.scan

pb.scan:
	@protoc --go_out=. hub3/ead/eadpb/scan.proto

pb.domain:
	@protoc --go_out=. ikuzo/domain/domainpb/domain.proto
	@protoc --go_out=. ikuzo/domain/domainpb/index.proto

pb.api:
	@protoc --go_out=. hub3/fragments/api.proto


pprof-dev:
	@pprof --http localhost:6060 -seconds 30 http://localhost:3000/debug/pprof/profile

# ikuzo specific make actions 
uncovered-ikuzo:
	richgo test -coverprofile /tmp/c.out ./... ; uncover /tmp/c.out

IKUZOMODULE:=github.com/delving/hub3/ikuzo/ikuzoctl

IKUZOLDFLAGS:=-X $(IKUZOMODULE)/cmd.version=`git describe --abbrev=0 --tags` -X $(IKUZOMODULE)/cmd.buildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X $(IKUZOMODULE)/cmd.gitHash=`git rev-parse HEAD` -X $(IKUZOMODULE)/cmd.buildAgent=`git config user.email`

build:
	go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

build-static:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

pre-commit:
	go mod tidy
	richgo test -cover -race -count=10 ./...
	golangci-lint run

test:
	richgo test -cover ./ikuzo/...
	golangci-lint run

test-no-cache:
	richgo test -cover -count=1 ./ikuzo/...
	golangci-lint run ikuzo

lint-full:
	golangci-lint run --enable=godox --enable=gomnd --enable=maligned --enable=prealloc --enable=gochecknoglobals --enable=gochecknoinits  ikuzo

run-dev:
	gin --path . --build ikuzo/ikuzoctl -i -buildArgs "-tags=dev -race -ldflags '${IKUZOLDFLAGS}'" run serve

generate-assets:
	go run ikuzo/internal/assets/generate.go

gen-oto:
	cp ikuzo/definitions/templates/def.d.ts frontend/gen/def.d.ts
	oto -template ikuzo/definitions/templates/def.js.plush -ignore Ignorer -pkg generated ikuzo/definitions/*.go > frontend/gen/def.js
	oto -template ikuzo/definitions/templates/clients.d.ts.plush -ignore Ignorer -pkg generated ikuzo/definitions/*.go > frontend/gen/clients.d.ts
	oto -template ikuzo/definitions/templates/clients.js.plush -ignore Ignorer -pkg generated ikuzo/definitions/*.go > frontend/gen/clients.js
	oto -template ikuzo/definitions/templates/server.go.plush -ignore Ignorer -pkg generated ikuzo/definitions/namespace.go   | gofmt \
		> ikuzo/definitions/generated/namespace.gen.go

