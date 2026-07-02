package main

import (
	"fmt"

	"log-analyzer/database"
	"log-analyzer/handler"
	"net/http"
)


func setupRoutes() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/upload", handler.UploadFile)
	http.HandleFunc("/files", handler.ListFiles)
	http.HandleFunc("/logs", handler.GetLogs)
	http.ListenAndServe(":8000", nil)
}

func main() {
	if err := database.Init(); err != nil {
		panic(err)
	}
	fmt.Println("Server running on :8000")
	setupRoutes()

}