package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	temp "github.itu.dk/sibh/temp/gRPC"
	"google.golang.org/grpc"
)

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 3000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &peer{
		id:      ownPort,
		value:   -1,
		clients: make(map[int32]temp.IncrementClient),
		ctx:     ctx,
	}

	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		port := int32(3000) + int32(i)

		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := temp.NewIncrementClient(conn)
		p.clients[port] = c
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.incrementAll()
	}
}

func (p *peer) Increment(ctx context.Context, req *temp.IncrementRequest) (*temp.IncrementReply, error) {
	fmt.Println("hello3")
	p.value++
	fmt.Println("hello4")
	rep := &temp.IncrementReply{Value: p.value}
	return rep, nil
}

func (p *peer) incrementAll() {
	fmt.Println("hello")
	incRequest := &temp.IncrementRequest{}
	fmt.Println("hell2")

	for id, client := range p.clients {
		fmt.Println("hell3")
		reply, err := client.Increment(p.ctx, incRequest)

		if err != nil {
			fmt.Println("something went wrong")
			fmt.Println(err)
		}

		fmt.Printf("Got reply from id %v: %v\n", id, reply.Value)
	}
}

type peer struct {
	temp.UnimplementedIncrementServer
	id      int32
	value   int32
	clients map[int32]temp.IncrementClient
	ctx     context.Context
}
