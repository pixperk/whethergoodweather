package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	advisorpb "github.com/pixperk/effinarounf/shared/proto/advisorpb"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := advisorpb.NewAdvisorServiceClient(conn)

	req := &advisorpb.AdvisorRequest{
		Cities: []*advisorpb.CityData{
			{Location: "New York"},
			{Location: "London"},
		},
	}

	fmt.Println("Getting weather advice (streaming)...")
	fmt.Println(strings.Repeat("=", 50))

	stream, err := client.StreamAdvice(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if resp.IsComplete {
			fmt.Println("\nAdvice complete!")
			break
		}

		fmt.Print(resp.Chunk)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
}
