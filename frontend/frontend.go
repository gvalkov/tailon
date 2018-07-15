// +build dev

package frontend

import "net/http"
import "github.com/shurcooL/httpfs/union"

var Assets http.FileSystem = union.New(
	map[string]http.FileSystem{
		"/dist":      http.Dir("frontend/dist"),
		"/templates": http.Dir("frontend/templates"),
	})
