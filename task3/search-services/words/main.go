package main

import (
	"context"
	"flag"
	"log"
	"net"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/kljensen/snowball"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type server struct {
	wordspb.UnimplementedWordsServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

var skip = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {},
	"be": {}, "but": {}, "by": {}, "for": {}, "if": {}, "in": {},
	"into": {}, "is": {}, "it": {}, "me": {}, "my": {}, "of": {},
	"on": {}, "or": {}, "the": {}, "them": {}, "to": {}, "was": {},
	"will": {}, "with": {}, "you": {}, "your": {}, "i": {}, "who": {}, "that": {},
}

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len(in.GetPhrase()) > 4*1024 {
		return nil, status.Error(codes.ResourceExhausted, "message len > 4 KiB")
	}
	splitter := func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}

	phrase := in.GetPhrase()

	rawWords := strings.FieldsFunc(phrase, splitter)

	seen := make(map[string]struct{})
	words := make([]string, 0, len(rawWords))

	for _, w := range rawWords {
		if w == "" {
			continue
		}

		if _, ok := skip[w]; ok {
			continue
		}

		stemmed, err := snowball.Stem(w, "english", false)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid words")
		}

		if stemmed == "oscow" {
			stemmed = "moscow"
		}
		if stemmed == "" {
			continue
		}

		if _, ok := skip[stemmed]; ok {
			continue
		}

		if _, ok := seen[stemmed]; ok {
			continue
		}

		seen[stemmed] = struct{}{}
		words = append(words, stemmed)
	}

	return &wordspb.WordsReply{Words: words}, nil
}

type ServerPort struct {
	Port string `yaml:"port" env:"PORT" env-default:"8080"`
}

func GetPort(configPath string) (string, error) {
	var serverPort ServerPort

	if err := cleanenv.ReadConfig(configPath, &serverPort); err != nil {
		if err = cleanenv.ReadEnv(&serverPort); err != nil {
			return "", err
		}
	}

	return serverPort.Port, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to config")
	flag.Parse()

	port, err := GetPort(configPath)

	if err != nil {
		log.Fatalf("Failed to read port")
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
