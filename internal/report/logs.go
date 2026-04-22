package report

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/shared"
)

func AppendJSONLine(path string, value any) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = file.Write(append(data, '\n'))
	return err
}

func ReadLogs(path string) ([]shared.LogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []shared.LogEntry{}, nil
		}
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var logs []shared.LogEntry
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry shared.LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			logs = append(logs, entry)
		}
	}
	return logs, scanner.Err()
}

func ReadHistory(path string) ([]shared.HistoryRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []shared.HistoryRecord{}, nil
		}
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var history []shared.HistoryRecord
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry shared.HistoryRecord
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			history = append(history, entry)
		}
	}
	return history, scanner.Err()
}

func ClearFile(path string) error {
	return os.WriteFile(path, []byte{}, 0o644)
}

func ExportLogs(path string, logs []shared.LogEntry, format string) (string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	outputPath := fmt.Sprintf("%s.%s", path, format)
	switch format {
	case "json":
		data, err := json.MarshalIndent(logs, "", "  ")
		if err != nil {
			return "", err
		}
		return outputPath, os.WriteFile(outputPath, data, 0o644)
	case "csv":
		file, err := os.Create(outputPath)
		if err != nil {
			return "", err
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()
		if err := writer.Write([]string{"timestamp", "level", "category", "target", "message"}); err != nil {
			return "", err
		}
		for _, entry := range logs {
			if err := writer.Write([]string{
				entry.Timestamp.Format(time.RFC3339),
				entry.Level,
				entry.Category,
				entry.Target,
				entry.Message,
			}); err != nil {
				return "", err
			}
		}
		return outputPath, nil
	case "txt":
		var builder strings.Builder
		for _, entry := range logs {
			builder.WriteString(fmt.Sprintf("[%s] %s %s %s - %s\n",
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				strings.ToUpper(entry.Level),
				entry.Category,
				entry.Target,
				entry.Message,
			))
		}
		return outputPath, os.WriteFile(outputPath, []byte(builder.String()), 0o644)
	default:
		return "", fmt.Errorf("formato invalido: %s", format)
	}
}

func PruneLogsOlderThan(path string, maxAge time.Duration) error {
	logs, err := ReadLogs(path)
	if err != nil {
		return err
	}
	if maxAge <= 0 {
		return nil
	}
	cutoff := time.Now().Add(-maxAge)
	var filtered []shared.LogEntry
	for _, entry := range logs {
		if entry.Timestamp.After(cutoff) {
			filtered = append(filtered, entry)
		}
	}
	dataFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dataFile.Close()
	for _, entry := range filtered {
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		if _, err := dataFile.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	return nil
}
