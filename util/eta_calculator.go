package util

import (
	"anileha/db"
	"math"
	"sync"
	"time"
)

const MinTimeDiffMs int64 = 1000
const SmoothingFactor float64 = 0.1

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
	timeDiff := now.Sub(c.startTime)
	if timeDiff.Milliseconds() < MinTimeDiffMs {
		return
	}
	speed := (newValue - c.lastValue) / float64(timeDiff)
	c.approxSpeed = c.approxSpeed*(1-SmoothingFactor) + speed*SmoothingFactor
	c.lastTime = now
	c.lastValue = newValue
}

func (c *EtaCalculator) getEta() float64 {
	if c.isUndefined {
		return math.NaN()
	}
	if !c.isStarted {
		return math.Inf(1)
	}
	return (c.endValue - c.lastValue) / c.approxSpeed
}

func (c *EtaCalculator) getElapsedTime() float64 {
	if !c.isStarted {
		return math.NaN()
	}
	if c.isFinished {
		return float64(c.lastTime.Sub(c.startTime).Milliseconds())
	}
	return float64(time.Since(c.startTime))
}

func (c *EtaCalculator) getProgressImpl() float64 {
	if !c.isStarted || c.isUndefined {
		return math.NaN()
	}
	if c.isFinished {
		return 1
	}
	return (c.lastValue - c.startValue) / (c.endValue - c.startValue)
}

func (c *EtaCalculator) GetProgress() db.Progress {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return db.Progress{
		Progress:    c.getProgressImpl(),
		Eta:         c.getEta(),
		TimeElapsed: c.getElapsedTime(),
	}
}
