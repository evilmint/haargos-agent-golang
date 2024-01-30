package statistics

import (
	"fmt"
	"sync"
	"time"
)

type Statistics struct {
	lock                     sync.RWMutex
	StartTime                time.Time
	failedRequestCount       int
	observationsSentCount    int
	dataSentInKB             int
	jobsProcessedCount       int
	lastSuccessfulConnection time.Time
	haAccessTokenSet         bool
	z2mSet                   bool
	zhaSet                   bool
	agentVersion             string
}

func NewStatistics() *Statistics {
	return &Statistics{
		StartTime:          time.Now(),
		jobsProcessedCount: 0,
	}
}

func (s *Statistics) GetUptime() string {
	currentTime := time.Now()
	uptimeDuration := currentTime.Sub(s.StartTime)

	hours := int(uptimeDuration.Hours())
	minutes := int(uptimeDuration.Minutes()) % 60
	seconds := int(uptimeDuration.Seconds()) % 60

	return fmt.Sprintf("Uptime: %d hours, %d minutes, %d seconds", hours, minutes, seconds)
}

func (s *Statistics) GetFailedRequestCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.failedRequestCount
}

func (s *Statistics) IncrementFailedRequestCount() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.failedRequestCount++
}

func (s *Statistics) GetObservationsSentCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.observationsSentCount
}

func (s *Statistics) IncrementObservationsSentCount() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.observationsSentCount++
}

func (s *Statistics) GetDataSentInKB() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.dataSentInKB
}

func (s *Statistics) AddDataSentInKB(data int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.dataSentInKB += data
}

func (s *Statistics) GetAgentVersion() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.agentVersion
}

func (s *Statistics) SetAgentVersion(value string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.agentVersion = value
}

func (s *Statistics) GetHAAccessTokenSet() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.haAccessTokenSet
}

func (s *Statistics) SetHAAccessTokenSet(value bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.haAccessTokenSet = value
}

func (s *Statistics) GetZ2MSet() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.z2mSet
}

func (s *Statistics) SetZ2MSet(value bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.z2mSet = value
}

func (s *Statistics) GetZHASet() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.zhaSet
}

func (s *Statistics) SetZHASet(value bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.zhaSet = value
}

func (s *Statistics) GetLastSuccessfulConnection() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.lastSuccessfulConnection
}

func (s *Statistics) SetLastSuccessfulConnection(lastConn time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.lastSuccessfulConnection = lastConn
}

func (s *Statistics) IncrementJobsProcessedCount() {
	s.lock.RLock()
	defer s.lock.RUnlock()

	s.jobsProcessedCount += 1
}

func (s *Statistics) GetJobsProcessedCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.jobsProcessedCount
}
