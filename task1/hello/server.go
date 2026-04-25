package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
)

type HelloPort struct {
	Port string `yaml:"port" env:"HELLO_PORT"`
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong\n"))
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("empty name\n"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hello, %s!\n", name)))
}

func CreateServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", PingHandler)
	mux.HandleFunc("GET /hello", HelloHandler)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return &server
}

func GetPort(configPath string) string {
	var helloPort HelloPort

	err := cleanenv.ReadConfig(configPath, &helloPort)

	if err != nil {
		return "8080"
	}

	return helloPort.Port

}

func main() {
	configPath := flag.String("config", "", "path to config")
	flag.Parse()

	port := GetPort(*configPath)

	server := CreateServer(port)

	server.ListenAndServe()
}
