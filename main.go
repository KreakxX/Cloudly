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
		return
	}

	fileBytes, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, "Error while reading file", http.StatusInternalServerError)
	}

	print(fileBytes)

	db, err := sql.Open("mysql", "root:1234@cloudly")

	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

}
