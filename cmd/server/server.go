package main

import (
	"context"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	csspb "github.com/google/codesearch/proto/codesearch_service"
	inpb "github.com/google/codesearch/proto/index"
	srpb "github.com/google/codesearch/proto/search"
)

var (
	listen = flag.String("listen", ":2633", "Address and port to listen on")
)

type codesearchServer struct {
	//	csspb.UnimplementedCodesearchServiceServer
}

func New() *codesearchServer {
	return &codesearchServer{}
}
func (css *codesearchServer) Index(ctx context.Context, req *inpb.IndexRequest) (*inpb.IndexResponse, error) {
	log.Printf("Index RPC")
	return &inpb.IndexResponse{}, nil
}
func (css *codesearchServer) Search(ctx context.Context, req *srpb.SearchRequest) (*srpb.SearchResponse, error) {
	log.Printf("Search RPC")
	return &srpb.SearchResponse{}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	css := New()

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	csspb.RegisterCodesearchServiceServer(grpcServer, css)

	log.Printf("Codesearch server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
