.PHONY: package

NAME:=hub3
MAINTAINER:="Sjoerd Siebinga <sjoerd@delving.eu>"
DESCRIPTION:="Hub3: Linked Open Data Platform"
MODULE:=github.com/delving/hub3

GO ?= go
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'grep "Version = \"" hub3/hub3ctl/cmd/root.go  | cut -d\" -f2')
GOVERSION:=$(shell sh -c 'go version | cut -d " " -f3')

LDFLAGS:=-X hub3.hub3ctl.Version=123 -X hub3.hub3ctl.BuildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X hub3.hub3ctl.GitHash=`git rev-parse HEAD` -X hub3.hub3ctl.BuildAgent=`git config user.email`
IKUZOLDFLAGS:=-X ikuzo.ikuzoctl.Version=$(VERSION) -X ikuzo.ikuzoctl.BuildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X ikuzo.ikuzoctl.GitHash=`git rev-parse HEAD` -X ikuzo.ikuzoctl.BuildAgent=`git config user.email`

# var print rule
print-%  : ; @echo $* = $($*)

clean:
	rm -rf $(NAME) build report gin-bin result.bin *.coverprofile */*.coverprofile hub3/hub3.db hub3/models/hub3.db dist server/assets/assets_vfsdata.go

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
	@richgo test --bench=. -benchmem ./...

ginkgo:
	@ginkgo -r  -skipPackage go_tests

twatch:
	@ginkgo watch -r -skipPackage go_tests

docker-image:
	docker build -t $(NAME) .

docker-start:
	docker run -p 3001:3001 -d $(NAME)

docker-stop:
	@sh -c "docker ps -a -q --filter ancestor=$(NAME) | xargs docker stop "

docker-remove:
	@make docker-stop
	@sh -c "docker image list -q hub3| xargs docker image rm -f"

docker-clean-build:
	@make docker-remove; 
	@make docker-image; 
	@make docker-start; 
	docker ps -all

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

setup-npm:
	# used for getting dependencies to render swagger specifications
	@npm install

release:
	@make create-assets
	@goreleaser --rm-dist --skip-publish
	@rpm --addsign dist/*.rpm
	@debsigs --sign=origin -k E2D6BD239452B1ED15CB99A66C417F6E7521731E dist/*.deb

release-dirty:
	@make create-assets
	@goreleaser --rm-dist --skip-publish --snapshot --skip-validate
	@rpm --addsign dist/*.rpm

release-snapshot:
	@make create-assets
	@goreleaser --rm-dist --skip-publish --snapshot
	@rpm --addsign dist/*.rpm

release-public:
	@make create-assets
	@goreleaser --rm-dist --skip-publish

protobuffer:
	@make pb.api
	@make pb.viewconfig
	@make pb.webresource

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

pre-commit:
	go mod tidy
	richgo test -cover -race -count=10 ./...
	golangci-lint run

test-ikuzo:
	richgo test -cover ./...
	golangci-lint run
	
test-no-cache:
	richgo test -cover -count=1 ./ikuzo/...
	golangci-lint run ikuzo

lint-full-ikuzo:
	golangci-lint run --enable=godox --enable=gomnd --enable=maligned --enable=prealloc --enable=gochecknoglobals --enable=gochecknoinits  ikuzo

api-console:
	api-console build -t "RAML 1.0" -a docs/ikuzo/raml/api.raml -o static/api-console

run-dev-ikuzo:
	gin --path . --build ikuzo -i -buildArgs "-tags=dev -ldflags '${LDFLAGS}'" run serve

ikuzo-generate-assets:
	go run internal/assets/generate.go

