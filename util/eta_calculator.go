package util

import (
	"sync"
	"time"
)

const MinTimeDiffMs float64 = 1
const SmoothingFactor float64 = 0.2

// Progress decided not to use floats here cause their serialization is error-prone in Golang
type Progress struct {
	Progress int `json:"progress"`
	Elapsed  int `json:"elapsed"`
	Eta      int `json:"eta"`
	Speed    int `json:"speed"`
}

type EtaCalculator struct {
	mutex       sync.Mutex
	approxSpeed float64
	startValue  float64
	lastValue   float64
	endValue    float64
	lastTime    time.Time
	startTime   time.Time
	isStarted   bool
	isFinished  bool
	isUndefined bool
}

func NewEtaCalculator(startValue float64, endValue float64) *EtaCalculator {
	return &EtaCalculator{
		startValue: startValue,
		lastValue:  startValue,
		endValue:   endValue,
	}
}

func NewUndefinedEtaCalculator() *EtaCalculator {
	return &EtaCalculator{
		lastValue:   0,
		endValue:    0,
		isUndefined: true,
	}
}

func (c *EtaCalculator) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isStarted = true
	c.startTime = time.Now()
	c.lastTime = c.startTime
}

func (c *EtaCalculator) ContinueWithNewValues(startValue float64, endValue float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.approxSpeed = 0
	c.startValue = startValue
	c.endValue = endValue
	c.lastValue = startValue
	c.lastTime = time.Now()
}

func (c *EtaCalculator) Update(newValue float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.isUndefined {
		return
	}
	if !c.isStarted {
		return
	}
	now := time.Now()
	if newValue == c.endValue {
		c.isFinished = true
		c.lastTime = now
		return
	}
	timeDiff := now.Sub(c.lastTime)
	if timeDiff.Seconds() < MinTimeDiffMs {
		return
	}
	speed := (newValue - c.lastValue) / timeDiff.Seconds()
	c.approxSpeed = c.approxSpeed*(1-SmoothingFactor) + speed*SmoothingFactor
	c.lastTime = now
	c.lastValue = newValue
}

func (c *EtaCalculator) getEta() float64 {
	if c.isUndefined {
		return -1
	}
	if !c.isStarted {
		return -1
	}
	if c.approxSpeed == 0 {
		return -1
	}
	return (c.endValue - c.lastValue) / c.approxSpeed
}

func (c *EtaCalculator) getElapsedTime() float64 {
	if !c.isStarted {
		return -1
	}
	if c.isFinished {
		return c.lastTime.Sub(c.startTime).Seconds()
	}
	return time.Since(c.startTime).Seconds()
}

func (c *EtaCalculator) getProgressImpl() float64 {
	if !c.isStarted || c.isUndefined {
		return 0
	}
	if c.isFinished {
		return 1
	}
	return (c.lastValue - c.startValue) / (c.endValue - c.startValue)
}

func (c *EtaCalculator) GetProgress() Progress {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return Progress{
		Progress: int(100 * c.getProgressImpl()),
		Eta:      int(c.getEta()),
		Elapsed:  int(c.getElapsedTime()),
		Speed:    int(c.approxSpeed),
	}
}
