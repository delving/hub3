package elasticsearch

import (
	"context"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/matryer/is"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type elasticSuite struct {
	suite.Suite
	elasticSearchC testcontainers.Container
	ip             string
	port           nat.Port
	ctx            context.Context
}

func TestElasticSearchSuite(t *testing.T) {
	suite.Run(t, new(elasticSuite))
}

// nolint:gocritic
func (s *elasticSuite) SetupSuite() {
	is := is.New(s.T())

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

	s.ctx = context.Background()

	var err error
	s.elasticSearchC, err = testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	is.NoErr(err)

	s.ip, err = s.elasticSearchC.Host(s.ctx)
	is.NoErr(err)

	s.port, err = s.elasticSearchC.MappedPort(s.ctx, "9200")
	is.NoErr(err)
}

// nolint:gocritic
func (s *elasticSuite) TearDownSuite() {
	is := is.New(s.T())
	err := s.elasticSearchC.Terminate(s.ctx)
	is.NoErr(err)
}
