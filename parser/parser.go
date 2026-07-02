package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log-analyzer/models"
)

func ParseLog(file interface{ Read([]byte) (int, error) }) ([]models.LogEntry, error) {
	var entries []models.LogEntry

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
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