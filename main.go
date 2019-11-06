package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(Root))
	mux.Handle("/", http.StripPrefix("/", neuter(fileServer)))

	mux.HandleFunc("/upload", upload)

	log.Print("hfm listen on " + strconv.Itoa(Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(Port), mux))
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.URL.Path)
		if r.URL.Path == "upload" {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/") || len(r.URL.Path) == 0 {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func upload(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Method)
	if r.Method != "POST" {
		w.WriteHeader(403)
		return
	}

	err := r.ParseMultipartForm(MaxDataSize)
	if err != nil {
		w.WriteHeader(403)
		return
	}

	filePath := r.FormValue("filePath")
	if filePath == "" {
		w.WriteHeader(403)
		return
	}

	fileData, handler, err := r.FormFile("fileData")
	if err != nil {
		w.WriteHeader(403)
		return
	}
	defer fileData.Close()

	fullPath := strings.Split(filePath, "/")
	path := strings.Join(fullPath[:len(fullPath)-1], "/")

	err = os.MkdirAll(Root+path, 0777)
	if err != nil {
		w.WriteHeader(503)
		return
	}

	f, err := os.OpenFile(Root+strings.Join(fullPath, "/"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(503)
		return
	}
	defer f.Close()

	if _, err = io.Copy(f, fileData); err != nil {
		w.WriteHeader(503)
		return
	}

	if _, err := fmt.Fprintf(w, "%v", handler.Header); err != nil {
		w.WriteHeader(200)
		return
	}
}
