package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cockroachdb/pebble"
	"github.com/google/codesearch/index"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	csspb "github.com/google/codesearch/proto/codesearch_service"
	inpb "github.com/google/codesearch/proto/index"
	srpb "github.com/google/codesearch/proto/search"
)

var (
	listen   = flag.String("listen", ":2633", "Address and port to listen on")
	indexDir = flag.String("index_dir", "", "Directory to store index in. Default: '~/.csindex/'")
)

type codesearchServer struct {
	db *pebble.DB
}

func defaultDir() string {
	var home string
	home = os.Getenv("HOME")
	if runtime.GOOS == "windows" && home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return filepath.Clean(home + "/.csindex")
}

func New() (*codesearchServer, error) {
	d := *indexDir
	if d == "" {
		d = defaultDir()
	}
	db, err := pebble.Open(d, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &codesearchServer{
		db: db,
	}, nil
}

func (css *codesearchServer) Index(ctx context.Context, req *inpb.IndexRequest) (*inpb.IndexResponse, error) {
	log.Printf("Index RPC")
	iw := index.Create(css.db)
	_ = iw
	return &inpb.IndexResponse{}, nil
}

func (css *codesearchServer) Search(ctx context.Context, req *srpb.SearchRequest) (*srpb.SearchResponse, error) {
	log.Printf("Search RPC")
	ir := index.Open(css.db)
	_ = ir
	return &srpb.SearchResponse{}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	css, err := New()
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	csspb.RegisterCodesearchServiceServer(grpcServer, css)

	log.Printf("Codesearch server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
