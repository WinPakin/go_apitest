package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/WinPakin/ackpb"

	"google.golang.org/grpc"
)

type server struct{}

func (*server) SendAck(ctx context.Context, req *ackpb.AckReq) (*ackpb.AckRes, error) {
	fmt.Printf("Greet function was invoked with %v\n", req)
	stamped := fmt.Sprintf("%s:rgpc-recv", req.Msg)
	res := &ackpb.AckRes{
		Msg: stamped,
	}
	return res, nil
}
func main() {
	fmt.Println("gRPC server running ...")
	lis, err := net.Listen("tcp", "0.0.0.0:5001")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	ackpb.RegisterAckServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
