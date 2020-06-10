// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
