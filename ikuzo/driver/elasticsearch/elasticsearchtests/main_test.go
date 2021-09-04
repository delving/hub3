package elasticsearchtests

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
)

var hostAndPort string

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run(
		"docker.elastic.co/elasticsearch/elasticsearch",
		"7.10.0",
		[]string{
			"discovery.type=single-node",
			"cluster.name=hub3_cluster",
			"xpack.security.enabled=false",
			"ES_JAVA_OPTS=-Xms1024m -Xmx1024m",
		})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort = resource.GetHostPort("9200/tcp")

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error

		resp, err := http.Get("http://" + hostAndPort)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("host not available: %s", hostAndPort)
		}

		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestSomething(t *testing.T) {
	// db.Query()
}
