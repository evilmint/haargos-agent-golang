package environmentgatherer

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/evilmint/haargos-agent-golang/repositories/commandrepository"
	"github.com/sirupsen/logrus"
)

type CPULoadManager struct {
	Logger       *logrus.Logger
	lastCPULoad  float64
	commandRepo  *commandrepository.CommandRepository
	stopFetching chan bool
	isFetching   bool
	mutex        sync.Mutex
}

func NewCPULoadManager(logger *logrus.Logger, commandRepo *commandrepository.CommandRepository) *CPULoadManager {
	manager := &CPULoadManager{
		Logger:       logger,
		commandRepo:  commandRepo,
		stopFetching: make(chan bool),
	}
	return manager
}

func (c *CPULoadManager) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Logger.Debugf("Fetch CPU Start (%p)", c)

	if !c.isFetching {
		c.isFetching = true
		go c.fetchPeriodically()
	}
}

func (c *CPULoadManager) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Logger.Debugf("Fetch CPU Stop")

	if c.isFetching {
		select {
		case c.stopFetching <- true:
		default:
			c.Logger.Error("Fetcher wasn't actively listening")
		}
		c.isFetching = false
	}
}

func (c *CPULoadManager) fetchPeriodically() {
	time.Sleep(time.Second)

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			c.Logger.Debugf("Fetch perdiocally CPU start")
			c.fetchCPULoad()
			c.Logger.Debugf("Fetch perdiocally CPU done")
		case <-c.stopFetching:

			c.Logger.Debugf("Got CPU fetch end event")
			return
		}
	}
}

func (c *CPULoadManager) fetchCPULoad() {
	top, err := c.commandRepo.GetCPULoad()
	if err != nil {
		c.Logger.Errorf("Error fetching CPU load: %v", err)
		return
	}

	load, err := strconv.ParseFloat(strings.TrimSpace(*top), 64)
	if err != nil {
		c.Logger.Errorf("Error parsing CPU load: %v", err)
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
