.PHONY: package

NAME:=rapid
MAINTAINER:="Sjoerd Siebinga <sjoerd@delving.eu>"
DESCRIPTION:="RAPID Linked Open Data Platform"
MODULE:=bitbucket.org/delving/rapid

GO ?= go
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'grep "Version = \"" cmd/root.go  | cut -d\" -f2')
GOVERSION:=$(shell sh -c 'go version | cut -d " " -f3')

LDFLAGS:="-X main.Version=$(VERSION) -X main.GoVersion=$(GOVERSION) -X main.BuildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X main.GitHash=`git rev-parse HEAD` -X main.BuildAgent=`git config user.email`"

# var print rule
print-%  : ; @echo $* = $($*)

install-glide:
	curl https://glide.sh/get | sh

clean:
	rm -rf $(NAME) build report gin-bin result.bin *.coverprofile */*.coverprofile

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
	gin run http

test:
	@go test  

ginkgo:
	@ginkgo -r 

twatch:
	@ginkgo watch -r

docker-image:
	gox -os="linux" -arch="amd64" -ldflags=$(LDFLAGS) -output="build/$(NAME)-{{.OS}}-{{.Arch}}" $(MODULE) 
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

goreport:
	@mkdir -p report
	@rm -rf report/*
	@goreporter -p ../rapid -r report -e vendor -f html
