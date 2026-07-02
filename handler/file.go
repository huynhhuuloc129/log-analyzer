package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"log-analyzer/database"
	"log-analyzer/parser"
	"net/http"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(35 << 20) // Limit upload size to 35MB

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "Error retrieving the file: %v", err)
		return
	}
	defer file.Close()

	// save file locally
	if err := SaveFile(file, handler.Filename, "uploads"); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Seek back to start so parseLog can read it again
	file.Seek(0, io.SeekStart)

	entries, err := parser.ParseLog(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing log: %v", err), http.StatusInternalServerError)
		return
	}

	// Save to SQLite
	if err := database.InsertEntries(handler.Filename, entries); err != nil {
		http.Error(w, fmt.Sprintf("Error saving to DB: %v", err), http.StatusInternalServerError)
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

func SaveFile(file interface{ Read([]byte) (int, error) }, filename string, foldername string) error {
	// Create the file on disk

	if err := os.MkdirAll("./"+foldername, os.ModePerm); err != nil {
		return err
	}

	dst, err := os.Create(fmt.Sprintf("./%s/%s", foldername, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy contents into it
	_, err = io.Copy(dst, file)
	return err
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`SELECT id, filename, uploaded_at FROM files ORDER BY uploaded_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type FileInfo struct {
		ID         int    `json:"id"`
		Filename   string `json:"filename"`
		UploadedAt string `json:"uploaded_at"`
	}

	var files []FileInfo
	for rows.Next() {
		var f FileInfo
		rows.Scan(&f.ID, &f.Filename, &f.UploadedAt)
		files = append(files, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func GetLogs(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("file_id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	page := r.URL.Query().Get("page")

	if fileID == "" {
		http.Error(w, "file_id is required", http.StatusBadRequest)
		return
	}

	if page == "" {
		page = "0"
	}

	query := `SELECT timestamp, level, message, source FROM entries WHERE file_id = ?`
	args := []interface{}{fileID}

	if from != "" {
		query += ` AND timestamp >= ?`
		args = append(args, from)
	}
	if to != "" {
		query += ` AND timestamp <= ?`
		args = append(args, to)
	}

	query += ` ORDER BY timestamp ASC LIMIT 5000 OFFSET ?`
	args = append(args, page)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type LogRow struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"level"`
		Message   string `json:"message"`
		Source    string `json:"source"`
	}

	var logs []LogRow
	for rows.Next() {
		var l LogRow
		var ts string
		rows.Scan(&ts, &l.Level, &l.Message, &l.Source)
		t, err := time.Parse("2006-01-02 15:04:05.9999999 -0700 -0700", ts)
		if err != nil {
			l.Timestamp = ts // fallback to raw string if parse fails
		} else {
			l.Timestamp = t.UTC().Format("2006-01-02 15:04:05.000")
		}
		logs = append(logs, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}