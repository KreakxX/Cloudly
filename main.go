package main

import (
	"database/sql"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	http.ListenAndServe(":8080", nil)
	http.HandleFunc("/upload", uploadFile)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	var file = r.Body

	if file == nil {
		return
	}

	db, err := sql.Open("mysql", "root:1234@cloudly")

	if err != nil {
		panic(err)
	}

	db.Query("SELECT * FROM cloudly")

}
