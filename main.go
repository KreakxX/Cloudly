package main

import (
	"database/sql"
	"io"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Chunk struct {
	Index int
	Data  []byte
}

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
	fileId := 10
	uploadChunksWithWorker(chunks, fileId)

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

func uploadChunksWithWorker(chunks [][]byte, fileId int) {
	db, err := sql.Open("mysql", "root:1234@cloudly")
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err != nil {
		panic(err)
	}

	chunkChan := make(chan Chunk)

	for w := 0; w < 8; w++ {
		go func(id int) {
			for chunk := range chunkChan {
				uploadEachChunk(db, chunk.Data, chunk.Index, fileId)
			}
		}(w)
	}

	// send after workers exist

	for i, data := range chunks {
		chunkChan <- Chunk{Index: i, Data: data}
	}

	close(chunkChan)

}

func uploadEachChunk(db *sql.DB, chunk []byte, index int, fileId int) error {
	stmt, err := db.Prepare("INSERT INTO file_chunks (file_id, chunk_index, data) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(fileId, index, chunk)
	if err != nil {
		return err
	}

	return nil
}
