package benchmark

import (
	"fmt"
	"testing"
)

func Test_parseLogFile(t *testing.T) {
	filePath := "../tmp/logfile.log"

	logEntries, err := parseLogFile(filePath)
	if err != nil {
		fmt.Println("Error parsing log file:", err)
		return
	}

	// 打印解析后的日志信息
	for _, entry := range logEntries {
		fmt.Printf("Timestamp: %s\n", entry.Timestamp)
		fmt.Printf("Level: %s\n", entry.Level)
		fmt.Printf("File: %s\n", entry.File)
		fmt.Printf("Line: %d\n", entry.Line)
		fmt.Printf("Message: %s\n", entry.Message)
		fmt.Println("----------")
	}
}
