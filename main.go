package main

import (
	"database/sql"
	"io"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	http.ListenAndServe(":8080", nil)
	http.HandleFunc("/upload", uploadFile)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Error while parsing File, might be too large", http.StatusInternalServerError)
	}

	file, header, err := r.FormFile("upload")

	print(header)

	if err != nil {
		http.Error(w, "Error while grabbing File contents", http.StatusInternalServerError)
	}

	fileBytes, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, "Error while reading file", http.StatusInternalServerError)
	}

	chunks := chunkFiles(fileBytes)

	if chunks == nil {
		http.Error(w, "Error while splitting file into Chunks", http.StatusInternalServerError)
	}

	uploadChunksWithWorker(chunks)

}

func chunkFiles(bytes []byte) [][]byte {
	var chunks [][]byte
	var chunkSize int = 4 * 1024 * 1024

	for i := 0; i < len(bytes); i += chunkSize {
		end := i + chunkSize

		if end > len(bytes) {
			end = len(bytes)
		}

		chunks = append(chunks, bytes[i:end])
	}

	return chunks
}

func uploadChunksWithWorker(chunks [][]byte) {

	chunkChan := make(chan []byte)

	for w := 0; w < 8; w++ {
		go func(id int) {
			for chunk := range chunkChan {
				uploadEachChunk(chunk)
			}
		}(w)
	}

	// send after workers exist

	for _, chunk := range chunks {
		chunkChan <- chunk
	}

	close(chunkChan)

}

func uploadEachChunk(chunk []byte) {
	db, err := sql.Open("mysql", "root:1234@cloudly")

	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}
