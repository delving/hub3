package index

import (
	"context"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/matryer/is"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type indexSuite struct {
	suite.Suite
	stanC testcontainers.Container
	ip    string
	port  nat.Port
	ctx   context.Context
}

func TestIndexSuite(t *testing.T) {
	suite.Run(t, new(indexSuite))
}

// nolint:gocritic
func (s *indexSuite) SetupSuite() {
	is := is.New(s.T())

	req := testcontainers.ContainerRequest{
		Image:        "nats-streaming:0.17.0",
		ExposedPorts: []string{"4222"},
		WaitingFor:   wait.ForLog("Streaming Server is ready"),
		Cmd: []string{
			"--cluster_id",
			"hub3-nats",
			"--http_port",
			"8222",
			"--port",
			"4222",
			"--max_bytes",
			"1GB",
			"--max_msgs",
			"1000000",
			// debugging information
			// "-SV",
			// "-SD",
		},
	}

	s.ctx = context.Background()

	var err error
	s.stanC, err = testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	is.NoErr(err)

	s.ip, err = s.stanC.Host(s.ctx)
	is.NoErr(err)

	s.port, err = s.stanC.MappedPort(s.ctx, "4222")
	is.NoErr(err)
}

// nolint:gocritic
func (s *indexSuite) TearDownSuite() {
	is := is.New(s.T())
	err := s.stanC.Terminate(s.ctx)
	is.NoErr(err)
}
