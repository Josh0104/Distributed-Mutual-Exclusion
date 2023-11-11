package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
    
    pb "github.com/Josh0104/Distributed-Mutal-Exclusion/proto"
	"google.golang.org/grpc"
)

type server struct {
	nodeId      string
	requested   bool
	timestamp   int64
	replyCount  int
	allowAccess chan bool
	mutex       sync.Mutex
}

func (s *server) RequestAccess(ctx context.Context, req *pb.Request) (*pb.Reply, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.timestamp = max(s.timestamp, req.Timestamp) + 1

	if s.requested && (s.timestamp < req.Timestamp || (s.timestamp == req.Timestamp && s.nodeId < req.NodeId)) {
		// Queue the request
		go func() {
			s.allowAccess <- false
		}()
	} else {
		// Reply immediately
		go func() {
			s.allowAccess <- true
		}()
	}

	return &pb.Reply{Allowed: true}, nil
}

func (s *server) enterCriticalSection() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.requested = true
	s.timestamp++
	s.replyCount = 0

	for i := 0; i < len(nodes)-1; i++ {
		go func(i int) {
			client := connectToNode(nodes[i])
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			request := &pb.Request{
				NodeId:    s.nodeId,
				Timestamp: s.timestamp,
			}

			reply, err := client.RequestAccess(ctx, request)
			if err != nil {
				log.Fatalf("Error requesting access from %s: %v", nodes[i], err)
			}

			if reply.Allowed {
				s.replyCount++
			}
		}(i)
	}

	// Wait for replies
	for i := 0; i < len(nodes)-1; i++ {
		<-s.allowAccess
	}

	// Enter Critical Section
	fmt.Printf("[%s] Entering Critical Section\n", s.nodeId)

	// Exit Critical Section
	s.requested = false
	fmt.Printf("[%s] Exiting Critical Section\n", s.nodeId)
}

func connectToNode(node string) pb.MutualExclusionClient {
	conn, err := grpc.Dial(node, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", node, err)
	}

	return pb.NewMutualExclusionClient(conn)
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

var nodes = []string{"localhost:50051", "localhost:50052", "localhost:50053"}

func main() {
	port := "50051" // change port accordingly for each node

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	server := grpc.NewServer()
	pb.RegisterMutualExclusionServer(server, &pb.server{
		nodeId:      fmt.Sprintf("Node-%s", port),
		allowAccess: make(chan bool),
	})

	fmt.Printf("[%s] Node started\n", fmt.Sprintf("Node-%s", port))

	go server.Serve(lis)

	// Allow time for other nodes to start
	time.Sleep(2 * time.Second)

	// Demonstrate entering Critical Section
	pb.server.enterCriticalSection()
}
