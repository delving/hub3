// Copyright © 2019 Delving B.V. <info@delving.eu>
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

package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/delving/hub3/hub3/namespace"
	pb "github.com/delving/hub3/hub3/server/grpc/pb/namespacepb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "launches the hub3 grpc and http server.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
		serve()
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

const (
	port = ":50051"
)

// server is used to implement the namespacepb.NamespaceServer
type service struct{}

// SearchLabel implements the namespacepb.SearchLabel rpc call
func (s *service) SearchLabel(ctx context.Context, in *pb.SearchLabelRequest) (*pb.SearchLabelResponse, error) {
	log.Printf("Received: %v", in.Uri)
	svc, err := namespace.NewService(namespace.WithDefaults())
	if err != nil {
		return nil, err
	}
	label, err := svc.SearchLabel(in.Uri)
	if err != nil {
		return nil, err
	}
	return &pb.SearchLabelResponse{Label: label}, nil
}

func serve() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterNamespaceServer(s, &service{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
