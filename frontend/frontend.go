// +build dev

package frontend

import "net/http"

var Assets http.FileSystem = http.Dir("frontend/dist")
