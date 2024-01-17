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
	lastSuccessfulConnection time.Time
}

func NewStatistics() *Statistics {
	return &Statistics{
		StartTime: time.Now(),
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
