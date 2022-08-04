package util

import (
	"math"
	"sync"
	"time"
)

const MinTimeDiffMs float64 = 1
const SmoothingFactor float64 = 0.2

type Progress struct {
	Progress float32
	Elapsed  uint
	Eta      uint
	Speed    float64
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
		approxSpeed: math.NaN(),
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
		c.approxSpeed = 0
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
		return c.lastTime.Sub(c.startTime).Seconds()
	}
	return time.Since(c.startTime).Seconds()
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

func (c *EtaCalculator) GetProgress() Progress {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return Progress{
		Progress: float32(c.getProgressImpl()),
		Eta:      uint(c.getEta()),
		Elapsed:  uint(c.getElapsedTime()),
		Speed:    c.approxSpeed,
	}
}
