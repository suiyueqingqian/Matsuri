package nekoray_rpc

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

type server struct {
	LibcoreServiceServer
}

var last time.Time
var nekoray_debug bool

func (s *server) KeepAlive(ctx context.Context, in *EmptyReq) (*EmptyResp, error) {
	last = time.Now()
	return &EmptyResp{}, nil
}

func Main() {
	_token := flag.String("token", "", "")
	_port := flag.Int("port", 19810, "")
	_debug := flag.Bool("debug", false, "")
	flag.CommandLine.Parse(os.Args[2:])

	nekoray_debug = *_debug

	go func() {
		t := time.NewTicker(time.Second * 10)
		for {
			<-t.C
			if last.Add(time.Second * 10).Before(time.Now()) {
				fmt.Println("Exit due to inactive")
				os.Exit(0)
			}
		}
	}()

	// Libcore
	setupCore()

	// GRPC
	lis, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(*_port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	token := *_token
	if token == "" {
		os.Stderr.WriteString("Please set a token: ")
		s := bufio.NewScanner(os.Stdin)
		if s.Scan() {
			token = strings.TrimSpace(s.Text())
		}
	}
	if token == "" {
		fmt.Println("You must set a token")
		os.Exit(0)
	}

	auther := Authenticator{
		Token: token,
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(auther.Authenticate)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(auther.Authenticate)),
	)
	RegisterLibcoreServiceServer(s, &server{})

	log.Printf("neokray grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
