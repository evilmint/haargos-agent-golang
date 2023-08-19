package loggatherer

import (
	"bufio"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type LogGatherer struct{}

func (l *LogGatherer) GatherLogs(haConfigPath string) string {
	file, err := os.Open(haConfigPath + "home-assistant.log")
	if err != nil {
		log.Errorf("Error reading log file: %v", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var logLines []string
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) >= 3 && (parts[2] == "WARNING" || parts[2] == "ERROR") {
			logLines = append(logLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Error scanning log file: %v", err)
		return ""
	}

	if len(logLines) > 200 {
		logLines = logLines[len(logLines)-200:]
	}

	// Join them by newline
	logContent := strings.Join(logLines, "\n")
	return logContent
}
