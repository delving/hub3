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
	rm -rf $(NAME) build report gin-bin

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
