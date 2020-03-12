// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

var FileSystem http.FileSystem = http.Dir("../../static")

func main() {

	err := vfsgen.Generate(FileSystem, vfsgen.Options{
		PackageName:  "assets",
		BuildTags:    "!dev",
		VariableName: "FileSystem",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
