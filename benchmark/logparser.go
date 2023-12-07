package benchmark

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	File      string
	Line      int
	Message   string
}

type MessageEntry struct {
	Timestamp time.Time
	Type      int
	Sender    int
	TxId      int
	Message   string
}

//
//var matchMap = map[int]string{
//	REQUEST:         `(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\.\d{3}[+-]\d{4}`,
//	SENT_PREPREPARE: `Submit Request: {alice tx1}`,
//	PREPREPARED:     `1 got message <PrePrepare with view: 0, seq: 1, payload of 20 bytes, header: MEcCAQETABNAYmExMzdjZDRkMzcxZGMzZDA4NTVkMzRiYjcxZGI2OGZjMGVhMTMwYTBlZDQ2OThiNWY5NjU2MWU3MzAwZDRkMA==> from 1 with seq 1`,
//	PREPARED:        `2 collected 2 prepares from [1 3]`,
//	COMMITTED:       `2 collected 2 commits from [3 4]`,
//	DECIDED:         `Deciding on seq 9`,
//}
//
//const (
//	REQUEST = iota
//	SENT_PREPREPARE
//	PREPREPARED
//	PREPARED
//	COMMITTED
//	DECIDED
//	REPLY
//)

func parse(c *Configuration) {
	var log = make(map[int][]*LogEntry)
	logDir := configuration.Log.LogDir
	for id := 1; id < 5; id++ {
		logFilePath := filepath.Join(logDir,
			fmt.Sprintf("node%d.log", id))
		logEntries, _ := parseLogFile(logFilePath)
		log[id] = logEntries
	}

}

func parseLogLine(logLine string) (*LogEntry, error) {
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\.\d{3}[+-]\d{4}\s+(\S+)\s+([\w/.]+.go):(\d+)\s+(.*)`)

	match := re.FindStringSubmatch(logLine)

	if len(match) != 6 {
		return nil, fmt.Errorf("log line does not match expected pattern: %s", logLine)
	}

	timestamp, err := time.Parse("2006-01-02T15:04:05", match[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing timestamp: %s", err)
	}

	line, err := fmt.Sscanf(match[4], "%d", new(int))
	if err != nil {
		return nil, fmt.Errorf("error parsing line number: %s", err)
	}

	return &LogEntry{
		Timestamp: timestamp,
		Level:     match[2],
		File:      match[3],
		Line:      line,
		Message:   match[5],
	}, nil
}

func parseLogFile(filePath string) ([]*LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logEntries []*LogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		logLine := scanner.Text()
		entry, err := parseLogLine(logLine)
		if err != nil {
			fmt.Printf("Error parsing log line '%s': %s\n", logLine, err)
			continue
		}
		logEntries = append(logEntries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logEntries, nil
}

func parseLogEntries(logEntries []*LogEntry) ([]*LogEntry, error) {
	var parsedEntries []*LogEntry
	for _, entry := range logEntries {
		if entry.Message == "Starting node" {
			parsedEntries = append(parsedEntries, entry)
		}
	}
	return parsedEntries, nil
}
