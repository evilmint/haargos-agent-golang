package commandrepository

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type CommandRepository struct {
	Logger *logrus.Logger
}

func NewCommandRepository(logger *logrus.Logger) *CommandRepository {
	return &CommandRepository{
		Logger: logger,
	}
}

type CPUInfo struct {
	Architecture string
	Model        string
	MHz          string
	CPUCount     int
}

func (c *CommandRepository) executeCommand(cmd string) (*string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, err
	}
	result := strings.TrimSpace(string(out))
	return &result, nil
}

func (c *CommandRepository) GetCPULoad() (*string, error) {
	return c.executeCommand("top -bn 1 | awk 'NR == 3 {printf \"%.2f\", 100 - $8}'")
}

func readArchitecture() (*string, error) {
	bytes, err := os.ReadFile("/proc/sys/kernel/arch")
	if err != nil {
		return nil, err
	}

	str := strings.TrimSpace(string(bytes))
	return &str, nil
}

func readCurrentFrequency() (*float32, error) {
	bytes, err := os.ReadFile("/sys/devices/system/cpu/cpufreq/policy0/cpuinfo_cur_freq")
	if err != nil {
		return nil, err
	}

	str := strings.TrimSpace(string(bytes))
	freq, err := strconv.ParseFloat(str, 32) // Parse as float32
	if err != nil {
		return nil, err
	}

	freq32 := float32(freq)
	return &freq32, nil
}

func (c *CommandRepository) GetCPUInfo() (*CPUInfo, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cpuInfo := CPUInfo{}
	scanner := bufio.NewScanner(file)
	cpuCount := 0

	arch, err := readArchitecture()
	if err != nil {
		return nil, err
	}

	mHz, err := readCurrentFrequency()
	cpuInfo.Architecture = *arch

	if err != nil {
		c.Logger.Errorf("Failed to read CPU frequency: %s.", err)
	} else {
		cpuInfo.MHz = fmt.Sprintf("%.4f", *mHz/1000.0)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "processor") {
			cpuCount++
		}
		if strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "CPU part") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				cpuInfo.Model = decodeCPUModel(strings.TrimSpace(parts[1]))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	cpuInfo.CPUCount = cpuCount
	return &cpuInfo, nil
}

func decodeCPUModel(model string) string {
	cpuPartMap := map[string]string{
		"0x810": "ARM810",
		"0x920": "ARM920",
		"0x922": "ARM922",
		"0x926": "ARM926",
		"0x940": "ARM940",
		"0x946": "ARM946",
		"0x966": "ARM966",
		"0xa20": "ARM1020",
		"0xa22": "ARM1022",
		"0xa26": "ARM1026",
		"0xb02": "ARM11 MPCore",
		"0xb36": "ARM1136",
		"0xb56": "ARM1156",
		"0xb76": "ARM1176",
		"0xc05": "Cortex-A5",
		"0xc07": "Cortex-A7",
		"0xc08": "Cortex-A8",
		"0xc09": "Cortex-A9",
		"0xc0d": "Cortex-A12",
		"0xc0f": "Cortex-A15",
		"0xc0e": "Cortex-A17",
		"0xc14": "Cortex-R4",
		"0xc15": "Cortex-R5",
		"0xc17": "Cortex-R7",
		"0xc18": "Cortex-R8",
		"0xc20": "Cortex-M0",
		"0xc21": "Cortex-M1",
		"0xc23": "Cortex-M3",
		"0xc24": "Cortex-M4",
		"0xc27": "Cortex-M7",
		"0xc60": "Cortex-M0+",
		"0xd01": "Cortex-A32",
		"0xd03": "Cortex-A53",
		"0xd04": "Cortex-A35",
		"0xd05": "Cortex-A55",
		"0xd07": "Cortex-A57",
		"0xd08": "Cortex-A72",
		"0xd09": "Cortex-A73",
		"0xd0a": "Cortex-A75",
		"0xd13": "Cortex-R52",
		"0xd20": "Cortex-M23",
		"0xd21": "Cortex-M33",
		"0x516": "ThunderX2",
		"0xa10": "SA110",
		"0xa11": "SA1100",
		"0x0a0": "ThunderX",
		"0x0a1": "ThunderX 88XX",
		"0x0a2": "ThunderX 81XX",
		"0x0a3": "ThunderX 83XX",
		"0x0af": "ThunderX2 99xx",
		"0x000": "X-Gene",
		"0x00f": "Scorpion",
		"0x02d": "Scorpion",
		"0x04d": "Krait",
		"0x06f": "Krait",
		"0x201": "Kryo",
		"0x205": "Kryo",
		"0x211": "Kryo",
		"0x800": "Falkor V1/Kryo",
		"0x801": "Kryo V2",
		"0xc00": "Falkor",
		"0xc01": "Saphira",
		"0x001": "exynos-m1",
		"0x003": "Denver 2",
		"0x131": "Feroceon 88FR131",
		"0x581": "PJ4/PJ4b",
		"0x584": "PJ4B-MP",
		"0x200": "i80200",
		"0x210": "PXA250A",
		"0x212": "PXA210A",
		"0x242": "i80321-400",
		"0x243": "i80321-600",
		"0x290": "PXA250B/PXA26x",
		"0x292": "PXA210B",
		"0x2c2": "i80321-400-B0",
		"0x2c3": "i80321-600-B0",
		"0x2d0": "PXA250C/PXA255/PXA26x",
		"0x2d2": "PXA210C",
		"0x411": "PXA27x",
		"0x41c": "IPX425-533",
		"0x41d": "IPX425-400",
		"0x41f": "IPX425-266",
		"0x682": "PXA32x",
		"0x683": "PXA930/PXA935",
		"0x688": "PXA30x",
		"0x689": "PXA31x",
		"0xb11": "SA1110",
		"0xc12": "IPX1200",
	}

	modelName, exists := cpuPartMap[model]
	if !exists {
		return model
	}

	return modelName
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

func (c *CommandRepository) GetRXTXBytes(interfaceName string, dataType string) (*int, error) {
	if dataType != "rx" && dataType != "tx" {
		return nil, fmt.Errorf("Invalid dataType: %s", dataType)
	}

	cmd := "cat /sys/class/net/" + interfaceName + "/statistics/" + dataType + "_bytes"
	data, err := c.executeCommand(cmd)

	if err != nil {
		return nil, err
	}

	intData, convErr := strconv.Atoi(*data)
	if convErr != nil {
		return nil, convErr
	}

	return &intData, nil
}

func (c *CommandRepository) GetRXTXPackets(interfaceName string, dataType string) (*int, error) {
	if dataType != "rx" && dataType != "tx" {
		return nil, fmt.Errorf("Invalid dataType: %s", dataType)
	}

	cmd := "cat /sys/class/net/" + interfaceName + "/statistics/" + dataType + "_packets"
	data, err := c.executeCommand(cmd)

	if err != nil {
		return nil, err
	}

	intData, convErr := strconv.Atoi(*data)
	if convErr != nil {
		return nil, convErr
	}

	return &intData, nil
}
