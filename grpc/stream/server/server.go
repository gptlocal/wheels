package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"

	pb "github.com/gptlocal/wheels/grpc/stream"
)

type server struct {
	pb.UnimplementedFibonacciServer
}

func (s *server) Calculate(req *pb.FibonacciRequest, stream pb.Fibonacci_CalculateServer) error {
	number := int(req.GetNumber())
	a, b := int64(0), int64(1)

	for i := 0; i < number; i++ {
		res := &pb.FibonacciResponse{
			Result: a,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
		a, b = b, a+b
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFibonacciServer(grpcServer, &server{})

	fmt.Println("Server is running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
