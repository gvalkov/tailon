//go:build ignore

package main

import (
	"log"

	"github.com/gvalkov/tailon/frontend"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(frontend.Assets, vfsgen.Options{
		PackageName:  "frontend",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
