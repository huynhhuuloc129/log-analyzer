package database

import (
	"database/sql"
	"log-analyzer/models"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("sqlite", "./logs.db")
	if err != nil {
		return err
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS files (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		filename    TEXT UNIQUE,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS entries (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id       INTEGER REFERENCES files(id),
		timestamp     DATETIME,
		level         TEXT,
		message       TEXT,
		source        TEXT,
		request_path  TEXT,
		request_id    TEXT,
		connection_id TEXT,
		process_id    INTEGER,
		thread_id     INTEGER
	)`)

	if err != nil {
		return err
	}
	
	_, err = DB.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS entries_fts USING fts5(
		message,
		source
	)`)
	if err != nil {
		return err
	}

	return err
}

// InsertEntries goes here below...

func InsertEntries(filename string, entries []models.LogEntry) error {
	// Insert file record and get its ID
	res, err := DB.Exec(`INSERT INTO files(filename) VALUES(?) ON CONFLICT(filename) DO UPDATE SET uploaded_at=CURRENT_TIMESTAMP`, filename)
	if err != nil {
		return err
	}

	fileID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Delete old entries in case of re-upload
	DB.Exec(`DELETE FROM entries WHERE file_id = ?`, fileID)

	// Insert all entries in one transaction (fast)
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO entries(file_id, timestamp, level, message, source, request_path, request_id, connection_id, process_id, thread_id) VALUES(?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range entries {
		_, err := stmt.Exec(fileID, e.Timestamp, e.Level, e.Message, e.SourceContext, e.RequestPath, e.RequestId, e.ConnectionId, e.ProcessId, e.ThreadId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit entries first
	if err := tx.Commit(); err != nil {
		return err
	}

	// Then build FTS index from committed entries
	_, err = DB.Exec(`INSERT INTO entries_fts(rowid, message, source) 
		SELECT id, message, source FROM entries WHERE file_id = ?`, fileID)
	return err
}