//go:build dev

package main

import "net/http"
import "github.com/shurcooL/httpfs/union"

var FrontendAssets http.FileSystem = union.New(
	map[string]http.FileSystem{
		"/dist":      http.Dir("frontend/dist"),
		"/templates": http.Dir("frontend/templates"),
	})
