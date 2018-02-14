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

package mediamanager_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	c "bitbucket.org/delving/rapid/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webresource", func() {

	c.InitConfig()
	var tmpDir string
	var relTmpDir string

	writeFile := func(folder string, filename string, content string, mode os.FileMode) {
		path := filepath.Join(tmpDir, folder)
		err := os.MkdirAll(path, 0700)
		Ω(err).ShouldNot(HaveOccurred())

		log.Println(tmpDir)
		path = filepath.Join(path, filename)
		ioutil.WriteFile(path, []byte(content), mode)
	}

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("/tmp", "rapid")
		c.Config.WebResource.WebResourceDir = tmpDir
		Ω(err).ShouldNot(HaveOccurred())

		cwd, err := os.Getwd()
		Ω(err).ShouldNot(HaveOccurred())
		relTmpDir, err = filepath.Rel(cwd, tmpDir)
		Ω(err).ShouldNot(HaveOccurred())

		//go files in the root directory (no tests)
		writeFile("/", "main.go", "package main", 0666)

		////non-go files in a nested directory
		//writeFile("/redherring", "big_test.jpg", "package ginkgo", 0666)

		////non-ginkgo tests in a nested directory
		//writeFile("/professorplum", "professorplum_test.go", `import "testing"`, 0666)

		////ginkgo tests in a nested directory
		//writeFile("/colonelmustard", "colonelmustard_test.go", `import "github.com/onsi/ginkgo"`, 0666)

		////ginkgo tests in a deeply nested directory
		//writeFile("/colonelmustard/library", "library_test.go", `import "github.com/onsi/ginkgo"`, 0666)

		////ginkgo tests deeply nested in a vendored dependency
		//writeFile("/vendor/mrspeacock/lounge", "lounge_test.go", `import "github.com/onsi/ginkgo"`, 0666)

		////a precompiled ginkgo test
		//writeFile("/precompiled-dir", "precompiled.test", `fake-binary-file`, 0777)
		//writeFile("/precompiled-dir", "some-other-binary", `fake-binary-file`, 0777)
		//writeFile("/precompiled-dir", "nonexecutable.test", `fake-binary-file`, 0666)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("WebResource parser", func() {

		Context("when receiving an urn", func() {

			It("should create a WebResourceRequest", func() {

			})
		})
	})

})
