// go:generate echo hello go generate!

// Package assets contains a virtual http.FileSytem with embedded static assets for the WebServer.
// All static content from the 'static' folder in the root of the project is included.
//
// assets.FileSystem can be used as any other http.FileSystem in your program, for example
//
// http.Handle("/assets", http.FileServer(assets.FileSystem))
//
// You must run 'go generate ./...' to update the contents of the assets.
//
//
package assets
