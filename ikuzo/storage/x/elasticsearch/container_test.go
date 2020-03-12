// +build external

package elasticsearch

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestElasticSearchVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping elasticsearch e2e in short mode")
	}

	ctx := context.Background()
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
	elasticSearchC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Error(err)
		return
	}

	defer elasticSearchC.Terminate(ctx)

	ip, err := elasticSearchC.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	port, err := elasticSearchC.MappedPort(ctx, "9200")
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestNginxLatestReturn(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Error(err)
	}

	defer nginxC.Terminate(ctx)

	ip, err := nginxC.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	port, err := nginxC.MappedPort(ctx, "80")
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	if err != nil {
		t.Error(err)
	}
}
