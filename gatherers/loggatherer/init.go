package loggatherer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type LogGatherer struct{}

// GatherLogs retrieves the log entries with WARNING or ERROR levels.
// It returns the last 200 such lines as a single string.
func (l *LogGatherer) GatherLogs(haConfigPath string) string {
	logFile := haConfigPath + "home-assistant.log"
	lines, err := readLogLines(logFile)
	if err != nil {
		log.Errorf("Error reading log file: %v", err)
		return ""
	}

	logLines := filterLogLines(lines)
	logContent := strings.Join(logLines, "\n")
	return logContent
}

func readLogLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening log file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error scanning log file: %w", err)
	}

	return lines, nil
}

func filterLogLines(lines []string) []string {
	var logLines []string
	for _, line := range lines {
		parts := strings.Fields(line)

		if len(parts) >= 3 && (parts[2] == "WARNING" || parts[2] == "ERROR") {
			logLines = append(logLines, line)
		}
	}

	// Keep only the last 200 lines if there are more
	if len(logLines) > 100 {
		logLines = logLines[len(logLines)-100:]
	}

	return logLines
}
