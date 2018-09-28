// +build dev

package buildpack

import "net/http"

var Assets http.FileSystem = http.Dir("tmpl")
