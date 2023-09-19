package environmentgatherer

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/evilmint/haargos-agent-golang/repositories/commandrepository"
)

type CPULoadManager struct {
	lastCPULoad  float64
	commandRepo  *commandrepository.CommandRepository
	stopFetching chan bool
	isFetching   bool
	mutex        sync.Mutex
}

func NewCPULoadManager(commandRepo *commandrepository.CommandRepository) *CPULoadManager {
	manager := &CPULoadManager{
		commandRepo:  commandRepo,
		stopFetching: make(chan bool),
	}
	return manager
}

func (c *CPULoadManager) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isFetching {
		c.isFetching = true
		go c.fetchPeriodically()
	}
}

func (c *CPULoadManager) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isFetching {
		select {
		case c.stopFetching <- true:
		default:
			log.Error("Fetcher wasn't actively listening")
		}
		c.isFetching = false
	}
}

func (c *CPULoadManager) fetchPeriodically() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.fetchCPULoad()
		case <-c.stopFetching:
			return
		}
	}
}

func (c *CPULoadManager) fetchCPULoad() {
	top, err := c.commandRepo.GetCPULoad()
	if err != nil {
		log.Errorf("Error fetching CPU load: %v", err)
		return
	}

	load, err := strconv.ParseFloat(strings.TrimSpace(*top), 64)
	if err != nil {
		log.Errorf("Error parsing CPU load: %v", err)
		return
	}

	c.mutex.Lock()
	c.lastCPULoad = load
	c.mutex.Unlock()
}

func (c *CPULoadManager) GetLastCPULoad() float64 {
	c.mutex.Lock()
	load := c.lastCPULoad
	c.mutex.Unlock()
	return load
}
