.PHONY: package deploy deploy-arm build-static-arm

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

# Deployment variables
# Support both SSH_HOST and SSH_TARGET for flexibility
SSH_TARGET ?=
SSH_HOST ?= $(SSH_TARGET)
BUILD_PATH := build/ikuzoctl
TEMP_PATH := /tmp/ikuzoctl
REMOTE_PATH := /opt/hub3/ikuzoctl

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


benchmark:
	@go tool richgo test --bench=. -benchmem ./...

compose-up:
	@docker-compose up

compose-down:
	@docker-compose down

compose-clean:
	@docker-compose down --volumes

protobuffer:
	@make pb.api
	@make pb.domain
	@make pb.scan
	@make viewconfig.api

pb.scan:
	@protoc --go_out=. hub3/ead/eadpb/scan.proto

pb.domain:
	@protoc --go_out=. ikuzo/domain/domainpb/domain.proto
	@protoc --go_out=. ikuzo/domain/domainpb/index.proto

pb.api:
	@protoc --go_out=. hub3/fragments/api.proto

viewconfig.api:
	@protoc --go_out=. hub3/fragments/viewconfig.proto

pprof-dev:
	@pprof --http localhost:6060 -seconds 30 http://localhost:3000/debug/pprof/profile

# ikuzo specific make actions 
uncovered-ikuzo:
	go tool richgo test -coverprofile /tmp/c.out ./... ; uncover /tmp/c.out

build:
	go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

build-static:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

build-static-arm:
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/ikuzoctl -ldflags "$(IKUZOLDFLAGS)" ikuzo/ikuzoctl/main.go

staticcheck:
	go tool staticcheck -f stylish ./hub3/... ./config/... ./ikuzo/...


loc:
	gcloc --exclude=.idea --exclude=_scratch --exclude=data --exclude=docker-data .

pre-commit:
	go mod tidy
	go tool richgo test -cover -race -count=10 ./...
	go tool staticcheck

test:
	go tool richgo test -cover ./ikuzo/...
	@make staticcheck

test-no-cache:
	go tool richgo test -cover -count=1 ./ikuzo/...
	@make staticcheck

run-dev:
	@go tool air

run-workers:
	@go tool air -c .air.workers.toml

define deploy_ikuzo
	@make build-static
	@echo "Deploying ikuzoctl to $(SSH_HOST)..."
	rsync -avP $(BUILD_PATH) $(SSH_HOST):$(TEMP_PATH)
	ssh $(SSH_HOST) 'sudo systemctl stop ikuzoctl; sudo mv $(TEMP_PATH) $(REMOTE_PATH); sudo systemctl start ikuzoctl'
	@echo "Deployment complete!"
endef

define deploy_ikuzo_arm
	@make build-static-arm
	@echo "Deploying ikuzoctl (arm64) to $(SSH_HOST)..."
	rsync -avP $(BUILD_PATH) $(SSH_HOST):$(TEMP_PATH)
	ssh $(SSH_HOST) 'sudo systemctl stop ikuzoctl; sudo mv $(TEMP_PATH) $(REMOTE_PATH); sudo systemctl start ikuzoctl'
	@echo "Deployment complete!"
endef

deploy:
	$(call deploy_ikuzo)

deploy-arm:
	$(call deploy_ikuzo_arm)
