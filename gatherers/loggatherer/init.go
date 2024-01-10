package loggatherer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/evilmint/haargos-agent-golang/client"
	"github.com/sirupsen/logrus"
)

type LogGatherer struct {
	Logger *logrus.Logger
}

func NewLogGatherer(logger *logrus.Logger) *LogGatherer {
	return &LogGatherer{
		Logger: logger,
	}
}

// GatherCoreLogs retrieves the log entries with WARNING or ERROR levels.
// It returns the last 200 such lines as a single string.
func (l *LogGatherer) GatherCoreLogs(haConfigPath string) string {
	logFile := haConfigPath + "home-assistant.log"
	lines, err := readLogLines(logFile)
	if err != nil {
		l.Logger.Errorf("Error reading log file: %v", err)
		return ""
	}

	logLines := filterLogLines(lines)
	logContent := strings.Join(logLines, "\n")
	return logContent
}

func (l *LogGatherer) GatherSupervisorLogs(client *client.HaargosClient, supervisorToken string) (string, error) {
	logs, err := client.FetchText("supervisor/logs", map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)})

	if err != nil {
		return "", err
	}

	return logs, nil
}

func (l *LogGatherer) GatherHostLogs(client *client.HaargosClient, supervisorToken string) (string, error) {
	logs, err := client.FetchText("host/logs", map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)})

	if err != nil {
		return "", err
	}

	return logs, nil
}

func readLogLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening log file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var lines []string
	var line []byte
	var isPrefix bool

	for {
		line, isPrefix, err = reader.ReadLine()
		if err != nil {
			break
		}

		fullLine := append([]byte(nil), line...)
		for isPrefix {
			line, isPrefix, err = reader.ReadLine()
			if err != nil {
				break
			}
			fullLine = append(fullLine, line...)
		}
		if err != nil {
			break
		}

		lines = append(lines, string(fullLine))
	}

	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("Error reading log file: %w", err)
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
