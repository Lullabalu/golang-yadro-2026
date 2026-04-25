package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/dustinkirkland/golang-petname"
	"github.com/ilyakaznacheev/cleanenv"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	petnamepb "yadro.com/course/proto"
)

type server struct {
	petnamepb.UnimplementedPetnameGeneratorServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) Generate(_ context.Context, r *petnamepb.PetnameRequest) (*petnamepb.PetnameResponse, error) {
	words := r.GetWords()
	separator := r.GetSeparator()

	if words <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	name := petname.Generate(int(words), separator)

	response := petnamepb.PetnameResponse{
		Name: name,
	}

	return &response, nil
}

func (s *server) GenerateMany(r *petnamepb.PetnameStreamRequest, stream grpc.ServerStreamingServer[petnamepb.PetnameResponse]) error {
	words := r.GetWords()
	names := r.GetNames()
	separator := r.GetSeparator()

	if words <= 0 || names <= 0 {
		return status.Error(codes.InvalidArgument, "invalid argument")
	}

	for range names {
		name := petname.Generate(int(words), separator)
		response := petnamepb.PetnameResponse{
			Name: name,
		}

		if err := stream.Send(&response); err != nil {
			return err
		}
	}

	return nil
}

type ServerPort struct {
	Port string `yaml:"port" env:"PETNAME_GRPC_PORT" env-default:"1234"`
}

func GetPort(config string) (string, error) {
	var serverPort ServerPort

	if err := cleanenv.ReadConfig(config, &serverPort); err != nil {
		if err := cleanenv.ReadEnv(&serverPort); err != nil {
			return "", err
		}
	}

	return serverPort.Port, nil
}

func main() {
	var config string
	flag.StringVar(&config, "config", "config.yaml", "path to config")
	flag.Parse()

	serverPort, err := GetPort(config)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	listener, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
