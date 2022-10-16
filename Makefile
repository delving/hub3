.PHONY: package

NAME:=hub3
MAINTAINER:="Sjoerd Siebinga <sjoerd@delving.eu>"
DESCRIPTION:="Hub3 (v2): Linked Open Data Platform"
MODULE:=github.com/delving/hub3/ikuzoctl

GO ?= go
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'git describe --abbrev=0 --tags')
GOVERSION:=$(shell sh -c 'go version | cut -d " " -f3')

IKUZOMODULE:=github.com/delving/hub3/ikuzo/ikuzoctl

IKUZOLDFLAGS:=-X $(IKUZOMODULE)/cmd.version=`git describe --abbrev=0 --tags` -X $(IKUZOMODULE)/cmd.buildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X $(IKUZOMODULE)/cmd.gitHash=`git describe --match=NeVeRmAtCh --always --abbrev=40 --dirty` -X $(IKUZOMODULE)/cmd.buildAgent=`git config user.email`
#
# var print rule
print-%  : ; @echo $* = $($*)

clean:
	rm -rf build report result.bin *.coverprofile */*.coverprofile hub3/hub3.db hub3/models/hub3.db dist server/assets/assets_vfsdata.go

clean-harvesting:
	rm -rf *_ids.txt *_records.xml

clean-build:
	@make clean
	mkdir -p build

create-assets:
	@go generate ./...

# dev dependencies
install-dev:
	go install github.com/cortesi/modd/cmd/modd@latest
	go install github.com/kyoh86/richgo@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

benchmark:
	@richgo test --bench=. -benchmem ./...

compose-up:
	@docker-compose up

compose-down:
	@docker-compose down

compose-clean:
	@docker-compose down --volumes

# TODO: replace with buff
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

build:
	go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

build-static:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

staticcheck:
	staticcheck -f stylish ./hub3/... ./config/... ./ikuzo/...

pre-commit:
	go mod tidy
	richgo test -cover -race -count=10 ./...
	@staticcheck

test:
	richgo test -cover ./ikuzo/...
	@make staticcheck

test-no-cache:
	richgo test -cover -count=1 ./ikuzo/...
	@make staticcheck

run-dev:
	modd
