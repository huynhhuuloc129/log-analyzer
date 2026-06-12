package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"log-analyzer/models"
	"log-analyzer/utility"
)

func parseLog(file interface{ Read([]byte) (int, error) }) ([]models.LogEntry, error) {
	var entries []models.LogEntry

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	
	lineNum := 0;
	for scanner.Scan() {
		lineNum++;
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry models.LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			fmt.Printf("Skipping line %d: %v\n", lineNum, err)
			continue // skip bad lines, don't crash
		}
		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(35 << 20) // Limit upload size to 35MB

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "Error retrieving the file: %v", err)
		return
	}
	defer file.Close()

	// save file locally 
	if err := utility.SaveFile(file, handler.Filename, "uploads"); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Seek back to start so parseLog can read it again
	file.Seek(0, io.SeekStart)

	entries, err := parseLog(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing log: %v", err), http.StatusInternalServerError)
		return
	}

	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Level]++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"file":    handler.Filename,
		"entries": len(entries),
		"levels":  counts,
	})
}

func setupRoutes(){
	http.HandleFunc("/upload", uploadFile);
	http.ListenAndServe(":8000", nil)
}

func main() {
	println("Hello, World!");
	setupRoutes()

}