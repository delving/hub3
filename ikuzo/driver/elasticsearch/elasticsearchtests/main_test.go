package elasticsearchtests

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var hostAndPort string

func TestMain(m *testing.M) {
	ctx := context.Background()

	// pulls an image, creates a container based on it and waits until elasticsearch answers
	es, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.elastic.co/elasticsearch/elasticsearch:7.14.1",
			ExposedPorts: []string{"9200/tcp"},
			WaitingFor:   wait.ForHTTP("/").WithPort("9200/tcp"),
			Env: map[string]string{
				"discovery.type":         "single-node",
				"cluster.name":           "hub3_cluster",
				"xpack.security.enabled": "false",
				"ES_JAVA_OPTS":           "-Xms1024m -Xmx1024m",
			},
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Could not start elasticsearch: %s", err)
	}

	hostAndPort, err = es.PortEndpoint(ctx, "9200/tcp", "")
	if err != nil {
		log.Fatalf("Could not get elasticsearch endpoint: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := es.Terminate(ctx); err != nil {
		log.Fatalf("Could not terminate elasticsearch: %s", err)
	}

	os.Exit(code)
}
