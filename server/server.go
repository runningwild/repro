package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/runningwild/repro/simple"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	grpcPort = flag.Int("grpc-port", 10000, "Port to serve grpc on")
	httpPort = flag.Int("http-port", 10001, "Port to serve http on")
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	simple.RegisterSimpleServer(grpcServer, &simpleServer{})
	go grpcServer.Serve(lis)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})
	go http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil)
	select {}
}
