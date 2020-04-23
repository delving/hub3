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

package fragments_test

import (
	"context"
	fmt "fmt"

	"github.com/delving/hub3/config"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"testing"
)

func TestFragments(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fragments Suite")
}

var (
	ctx            = context.Background()
	elasticSearchC testcontainers.Container
	ip             string
	port           nat.Port
)

var _ = BeforeSuite(func() {
	// init configuration
	config.InitConfig()

	req := testcontainers.ContainerRequest{
		Image:        "docker.elastic.co/elasticsearch/elasticsearch:7.6.1",
		ExposedPorts: []string{"9200"},
		// WaitingFor:   wait.ForHTTP(":9200/"),
		WaitingFor: wait.ForLog("indices into cluster_state"),
		Env: map[string]string{
			"discovery.type":         "single-node",
			"cluster.name":           "ikuzo_cluster",
			"xpack.security.enabled": "false",
			"ES_JAVA_OPTS":           "-Xms1024m -Xmx1024m",
		},
	}

	var err error
	elasticSearchC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	Expect(err).NotTo(HaveOccurred())

	ip, err = elasticSearchC.Host(ctx)
	Expect(err).NotTo(HaveOccurred())

	port, err = elasticSearchC.MappedPort(ctx, "9200")
	Expect(err).NotTo(HaveOccurred())

	config.Config.ElasticSearch.Urls = []string{fmt.Sprintf("http://%s:%s", ip, port.Port())}
})

var _ = AfterSuite(func() {
	err := elasticSearchC.Terminate(ctx)
	Expect(err).NotTo(HaveOccurred())
})
