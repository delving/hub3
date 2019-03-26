// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mediamanager

import (
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	c "github.com/delving/hub3/config"
	elastic "github.com/olivere/elastic"
)

const (
	sourceDir    = "source"
	thumbnailDir = "thumbnail"
	deepzoomDir  = "deepzoom"
)

var files = make(map[[sha512.Size]byte]string)

func printFile(ignoreDirs []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}
		if info.IsDir() {
			dir := filepath.Base(path)
			for _, d := range ignoreDirs {
				if d == dir {
					return filepath.SkipDir
				}
			}
		}
		fmt.Println(path)
		fmt.Printf("%+v", info)
		return nil
	}
}

func checkDuplicate(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
		return nil
	}
	if info.IsDir() {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return nil
	}
	digest := sha512.Sum512(data)
	if v, ok := files[digest]; ok {
		fmt.Printf("%q is a duplicate of %q\n", path, v)
	} else {
		files[digest] = path
	}
	return nil
}

// IndexWebResources reindexes all webresources
func IndexWebResources(p *elastic.BulkProcessor) error {
	err := filepath.Walk(c.Config.WebResource.WebResourceDir, printFile([]string{}))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
