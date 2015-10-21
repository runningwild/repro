package main

import (
	"flag"
	"io"
	"log"
	"sync"
	"time"

	"github.com/runningwild/repro/simple"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	echo       = flag.Int("echo", -1, "number to echo")
	count      = flag.Int("count", -1, "number to count to")
	sum        = flag.Int("sum", -1, "sum up to this number")
	echoStream = flag.Int("echostream", -1, "double count up to this number")
	N          = flag.Int("n", 1, "number to run in parallel")
)

func doEcho(client simple.SimpleClient, num int) {
	req := &simple.Number{Num: int32(num)}
	log.Printf("Echo: %v", req)
	reply, err := client.Echo(context.Background(), req)
	if err != nil {
		log.Printf("Failed to echo: %v", err)
		return
	}
	log.Printf("Echo reply: %v", reply)
}

// printFeatures lists all the features within the given bounding Rectangle.
func doCount(client simple.SimpleClient, num int) {
	req := &simple.Number{Num: int32(num)}
	log.Printf("Count: %v", req)
	stream, err := client.Count(context.Background(), req)
	if err != nil {
		log.Printf("Failed to count: %v", err)
		return
	}
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Count stream failed: %v", err)
			return
		}
		log.Printf("Count reply: %v", reply)
	}
	log.Printf("Counting complete")
}

func doSum(client simple.SimpleClient, num int) {
	log.Printf("Sum: %d", num)
	stream, err := client.Sum(context.Background())
	if err != nil {
		log.Printf("Failed to Sum: %v", err)
		return
	}
	for i := 0; i < num; i++ {
		if err := stream.Send(&simple.Number{Num: int32(i)}); err != nil {
			log.Printf("Failed to sum: %v", err)
			return
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Failed to sum: %v", err)
		return
	}
	log.Printf("Sum up to %d: %v", num, reply)
}

func doEchoStream(client simple.SimpleClient, num int) {
	log.Printf("EchoStream: %v", num)
	stream, err := client.EchoStream(context.Background())
	if err != nil {
		log.Printf("Failed EchoStream: %v", err)
		return
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Failed EchoStream: %v", err)
				return
			}
			log.Printf("EchoStream received %v", in)
		}
	}()
	defer func() {
		stream.CloseSend()
		<-done
	}()
	for i := 0; i < num; i++ {
		if err := stream.Send(&simple.Number{Num: int32(i)}); err != nil {
			log.Printf("Failed to EchoStream: %v", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	for i := 0; i < *N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doItAll()
		}()
	}
	wg.Wait()
}

func doItAll() {
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := simple.NewSimpleClient(conn)
	for conn.State() != grpc.Ready {
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
	log.Printf("connection is ready")
	if *echo >= 0 {
		doEcho(client, *echo)
	}
	if *count >= 0 {
		doCount(client, *count)
	}
	if *sum >= 0 {
		doSum(client, *sum)
	}
	if *echoStream >= 0 {
		doEchoStream(client, *echoStream)
	}
}
