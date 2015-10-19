package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/runningwild/repro/simple"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10000, "Port to serve on")
)

type simpleServer struct{}

func (*simpleServer) Echo(ctx context.Context, req *simple.Number) (*simple.Number, error) {
	return &simple.Number{Num: req.Num}, nil
}
func (*simpleServer) Count(req *simple.Number, stream simple.Simple_CountServer) error {
	for i := int32(0); i < req.Num; i++ {
		if err := stream.Send(&simple.Number{Num: i}); err != nil {
			return err
		}
	}
	return nil
}
func (*simpleServer) Sum(stream simple.Simple_SumServer) error {
	reply := &simple.Number{}
	for {
		num, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		reply.Num += num.Num
	}
	return stream.SendAndClose(reply)
}
func (*simpleServer) EchoStream(stream simple.Simple_EchoStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		for i := int32(0); i < in.Num; i++ {
			if err := stream.Send(&simple.Number{i}); err != nil {
				return err
			}
		}
	}
}
func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	simple.RegisterSimpleServer(grpcServer, &simpleServer{})
	grpcServer.Serve(lis)
}
