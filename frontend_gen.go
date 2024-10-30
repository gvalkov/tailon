//go:build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
)

var FrontendAssets http.FileSystem = union.New(
	map[string]http.FileSystem{
		"/dist":      http.Dir("frontend/dist"),
		"/templates": http.Dir("frontend/templates"),
	})

func main() {
	err := vfsgen.Generate(FrontendAssets, vfsgen.Options{
		PackageName:  "main",
		BuildTags:    "!dev",
		VariableName: "FrontendAssets",
		Filename:     "frontend_bin.go",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
