package commandrepository

import (
	"os/exec"
	"strings"
)

type CommandRepository struct{}

func (c *CommandRepository) executeCommand(cmd string) (*string, error) {
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, err
	}
	result := strings.TrimSpace(string(out))
	return &result, nil
}

func (c *CommandRepository) GetCPULoad() (*string, error) {
	return c.executeCommand("top -bn 1 | awk 'NR == 3 {printf \"%.2f\", 100 - $8}'")
}

func (c *CommandRepository) GetCPUInfo() (*string, error) {
	return c.executeCommand("lscpu | grep -E 'Architecture|Model name|CPU MHz|CPU(s)' | sed 's/   *//g'")
}

func (c *CommandRepository) GetStorage() (*string, error) {
	return c.executeCommand("df -h")
}

func (c *CommandRepository) GetMemory() (*string, error) {
	return c.executeCommand("free")
}

func (c *CommandRepository) GetLastBootTime() (*string, error) {
	return c.executeCommand("uptime -s")
}

func (c *CommandRepository) GetCPUTemperature() (*string, error) {
	return c.executeCommand("cat /sys/class/thermal/thermal_zone0/temp | awk '{printf \"%.1f\", $1/1000}'")
}

func (c *CommandRepository) GetNetworkInterfaces() (*string, error) {
	return c.executeCommand("ls /sys/class/net/")
}

func (c *CommandRepository) GetRXBytes(interfaceName string) (*string, error) {
	return c.executeCommand("cat /sys/class/net/" + interfaceName + "/statistics/rx_bytes")
}

func (c *CommandRepository) GetTXBytes(interfaceName string) (*string, error) {
	return c.executeCommand("cat /sys/class/net/" + interfaceName + "/statistics/tx_bytes")
}
