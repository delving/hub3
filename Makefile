.PHONY: package

NAME:=rapid
MAINTAINER:="Sjoerd Siebinga <sjoerd@delving.eu>"
DESCRIPTION:="RAPID Linked Open Data Platform"
MODULE:=bitbucket.org/delving/rapid

GO ?= go
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'grep "Version = \"" cmd/root.go  | cut -d\" -f2')
GOVERSION:=$(shell sh -c 'go version | cut -d " " -f3')

LDFLAGS:=-X main.Version=$(VERSION) -X main.BuildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X main.GitHash=`git rev-parse HEAD` -X main.BuildAgent=`git config user.email`

# var print rule
print-%  : ; @echo $* = $($*)

install-glide:
	curl https://glide.sh/get | sh

clean:
	rm -rf $(NAME) build report gin-bin result.bin *.coverprofile */*.coverprofile hub3/rapid.db hub3/models/rapid.db dist

clean-harvesting:
	rm -rf *_ids.txt *_records.xml

clean-build:
	@make clean
	mkdir -p build

run:
	@go run main.go

build:
	@make clean-build
	@go build -a -o build/$(NAME) -ldflags=$(LDFLAGS) $(MODULE)

gox-build:
	@make clean-build
	cd build 
	@make build 
	gox -os="linux" -os="darwin" -os="windows" -arch="amd64" -ldflags=$(LDFLAGS) -output="build/$(NAME)-{{.OS}}-{{.Arch}}" $(MODULE) 
	ls -la ./build/

run-dev:
	gin -buildArgs "-i -ldflags '${LDFLAGS}'" run http

test:
	@go test  ./...

benchmark:
	@go test --bench=. -benchmem ./...

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
	@sh -c "docker image list -q rapid | xargs docker image rm -f"

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
	@goreporter -p ../rapid -r report -e vendor -f html

setup-npm:
	# used for getting dependencies to render swagger specifications
	@npm install

release:
	@goreleaser --rm-dist --skip-publish
	@rpm --addsign dist/*.rpm
	@debsigs --sign=origin -k E2D6BD239452B1ED15CB99A66C417F6E7521731E dist/*.deb

release-dirty:
	@goreleaser --rm-dist --skip-publish --snapshot --skip-validate
	@rpm --addsign dist/*.rpm

release-snapshot:
	@goreleaser --rm-dist --skip-publish --snapshot
	@rpm --addsign dist/*.rpm

release-public:
	@goreleaser --rm-dist --skip-publish

protobuffer:
	@make pb.fragment
	@make pb.api
	@make pb.webresource

pb.webresource:
	@protoc --go_out=. hub3/mediamanager/webresource.proto

pb.fragment:
	@protoc --go_out=plugins=grpc:. hub3/fragments/fragments.proto

pb.api:
	@echo "correct the import path in the api.pb.go file for fragments."
	@protoc --go_out=. hub3/api/api.proto

elm:
	@elm-make elm-src/Main.elm --output=public/js/elm.js
