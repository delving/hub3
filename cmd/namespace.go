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
	"time"

	pb "github.com/delving/hub3/pkg/server/grpc/pb/namespacepb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// namespaceCmd represents the namespace command
var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "client for hub3 namespaces service.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("namespace called")
		searchLabel()
	},
}

func init() {
	RootCmd.AddCommand(namespaceCmd)
}

const (
	address     = "localhost:50051"
	defaultName = "http://purl.org/dc/elements/1.1/title"
)

func searchLabel() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewNamespaceClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	//if len(os.Args) > 1 {
	//name = os.Args[1]
	//}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SearchLabel(ctx, &pb.SearchLabelRequest{Uri: name})
	if err != nil {
		log.Fatalf("could not find namespace: %v", err)
	}
	log.Printf("namespace: %s", r.Label)
}
