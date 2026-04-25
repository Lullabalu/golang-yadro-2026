package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/ilyakaznacheev/cleanenv"
)

type FileServerPort struct {
	Port string `yaml:"port" env:"FILESERVER_PORT"`
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	existingFile, err := os.Open("./uploads/" + header.Filename)
	if err == nil {
		existingFile.Close()
		w.WriteHeader(http.StatusConflict)
		return
	}
	if !os.IsNotExist(err) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	out, err := os.Create("./uploads/" + header.Filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(header.Filename))
}

func PutHandler(w http.ResponseWriter, r *http.Request) {
	src, _, err := r.FormFile("file")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer src.Close()

	filename := r.PathValue("filename")

	file, err := os.OpenFile("./uploads/"+filename, os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	defer file.Close()
	file.Truncate(0)
	file.Seek(0, 0)

	defer r.Body.Close()
	io.Copy(file, src)

	w.WriteHeader(http.StatusOK)
}

func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir("./uploads")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	files := []string{}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	w.WriteHeader(http.StatusOK)

	for _, name := range files {
		w.Write([]byte(name))
		w.Write([]byte("\n"))
	}
}

func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	file, err := os.Open("./uploads/" + filename)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	defer file.Close()

	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	err := os.Remove("./uploads/" + filename)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func CreateServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /files", PostHandler)
	mux.HandleFunc("PUT /files/{filename}", PutHandler)
	mux.HandleFunc("GET /files", GetFilesHandler)
	mux.HandleFunc("GET /files/{filename}", GetFileHandler)
	mux.HandleFunc("DELETE /files/{filename}", DeleteHandler)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return &server
}

func GetPort(configPath string) string {
	var fileserverPort FileServerPort

	err := cleanenv.ReadConfig(configPath, &fileserverPort)

	if err != nil {
		return "8080"
	}
	if fileserverPort.Port == "" {
		return "8080"
	}
	return fileserverPort.Port

}

func main() {
	configPath := flag.String("config", "", "path to config")
	flag.Parse()

	os.MkdirAll("./uploads", 0755)

	port := GetPort(*configPath)

	server := CreateServer(port)

	server.ListenAndServe()
}
