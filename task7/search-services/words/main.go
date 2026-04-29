package main

import (
	"context"
	"flag"
	"log"
	"net"
	"strings"
	"unicode"

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

const maxMessageSize = 1024 * 1024

var stopWords = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "but": {},
	"of": {}, "in": {}, "on": {}, "at": {}, "to": {}, "for": {},
	"me": {}, "your": {}, "them": {}, "who": {}, "that": {},
	"is": {}, "are": {}, "was": {}, "were": {}, "be": {}, "been": {},
	"will": {}, "would": {}, "can": {}, "could": {}, "should": {},
	"i": {}, "you": {}, "he": {}, "she": {}, "it": {}, "we": {}, "they": {},

	"и": {}, "а": {}, "но": {}, "или": {}, "в": {}, "во": {},
	"на": {}, "с": {}, "со": {}, "к": {}, "ко": {}, "по": {},
	"за": {}, "из": {}, "у": {}, "о": {}, "об": {}, "от": {},
	"это": {}, "этот": {}, "эта": {}, "эти": {},
	"я": {}, "ты": {}, "он": {}, "она": {}, "оно": {}, "мы": {}, "вы": {}, "они": {},
	"буду": {}, "будешь": {}, "будет": {}, "будем": {}, "будете": {}, "будут": {},
}

func splitWords(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func detectLang(word string) string {
	for _, r := range word {
		if unicode.In(r, unicode.Cyrillic) {
			return "russian"
		}
	}

	return "english"
}

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	phrase := in.GetPhrase()

	if len([]byte(phrase)) > maxMessageSize {
		return nil, status.Error(codes.ResourceExhausted, "message is larger than 4 KiB")
	}
	words := splitWords(phrase)

	result := make([]string, 0)
	existed := make(map[string]struct{}, len(words))

	for _, word := range words {
		word = strings.ToLower(word)

		if _, ok := stopWords[word]; ok {
			continue
		}

		lang := detectLang(word)

		stemmed, err := snowball.Stem(word, lang, true)
		if err != nil {
			stemmed = word
		}

		if stemmed == "" {
			continue
		}
		if _, ok := existed[stemmed]; ok {
			continue
		}

		result = append(result, stemmed)
		existed[stemmed] = struct{}{}
	}

	return &wordspb.WordsReply{Words: result}, nil
}

type ServerPort struct {
	Port string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:":1234"`
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

	listener, err := net.Listen("tcp", port)
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
