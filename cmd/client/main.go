package main

import (
	"context"
	"log"

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
			{Location: "New York", State: "NY", Country: "US"},
			{Location: "London", Country: "UK"},
		},
	}

	resp, err := client.GetAdvice(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Advice: %s", resp.Advice)
}
