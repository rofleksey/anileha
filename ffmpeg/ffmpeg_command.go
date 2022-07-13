package ffmpeg

import (
	"anileha/db"
	"anileha/util"
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Command struct {
	mutex sync.Mutex
	opts  []option

	// immutable
	frameRegex     *regexp.Regexp
	numberOfFrames int
}

type option struct {
	key      string
	priority OptionPriority
	value    *string
}

type OptionPriority int

const (
	OptionBase       OptionPriority = 1
	optionInputFile  OptionPriority = 2
	OptionInput      OptionPriority = 3
	OptionOutput     OptionPriority = 4
	optionOutputFile OptionPriority = 5
)

func (c *Command) AddSingle(key string, optType OptionPriority) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	opt := newOption(key, optType, nil)
	c.opts = append(c.opts, opt)
}

func (c *Command) AddEscapedSingle(key string, optType OptionPriority) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	escapedKey := fmt.Sprintf("\"%s\"", key)
	opt := newOption(escapedKey, optType, nil)
	c.opts = append(c.opts, opt)
}

func (c *Command) AddKeyValue(key string, value string, optType OptionPriority) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	opt := newOption(key, optType, &value)
	c.opts = append(c.opts, opt)
}

func (c *Command) AddEscapedKeyValue(key string, value string, optType OptionPriority) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	escapedValue := fmt.Sprintf("\"%s\"", value)
	opt := newOption(key, optType, &escapedValue)
	c.opts = append(c.opts, opt)
}

func newOption(key string, priority OptionPriority, value *string) option {
	return option{key, priority, value}
}

func (o *option) String() string {
	if o.value != nil {
		return fmt.Sprintf("%s %s", o.key, *o.value)
	}
	return o.key
}

func NewCommand(inputFile string, numberOfFrames int, outputFile string) *Command {
	frameRegex := regexp.MustCompile("frame=\\s+?(\\d+)")
	command := Command{
		opts:           make([]option, 0, 32),
		frameRegex:     frameRegex,
		numberOfFrames: numberOfFrames,
	}
	command.AddSingle("-hide_banner", OptionBase)
	command.AddKeyValue("-hw_accel", "auto", OptionBase)
	command.AddEscapedKeyValue("-i", inputFile, optionInputFile)
	command.AddEscapedSingle(outputFile, optionOutputFile)
	return &command
}

func (c *Command) progressWatcher(reader io.ReadCloser, progressChan db.ProgressChan) {
	scanner := bufio.NewScanner(reader)
	var etaCalculator *util.EtaCalculator
	if c.numberOfFrames != 0 {
		etaCalculator = util.NewEtaCalculator(0, float64(c.numberOfFrames))
	} else {
		etaCalculator = util.NewUndefinedEtaCalculator()
	}
	etaCalculator.Start()
	for scanner.Scan() {
		line := scanner.Text()
		matchResult := c.frameRegex.FindStringSubmatch(line)
		if matchResult != nil {
			curFrame, _ := strconv.Atoi(matchResult[0])
			etaCalculator.Update(float64(curFrame))
			progress := etaCalculator.GetProgress()
			progressChan <- progress
		}
	}
}

func (c *Command) completionWatcher(cancelFunc context.CancelFunc, cmd *exec.Cmd, reader io.ReadCloser, finishChan db.FinishChan) {
	err := cmd.Wait()
	_ = reader.Close()
	cancelFunc()
	finishChan <- err
	close(finishChan)
}

func (c *Command) prepareArgs(withExecutable bool) []string {
	sort.SliceStable(c.opts, func(i, j int) bool {
		return c.opts[i].priority < c.opts[j].priority
	})
	args := make([]string, 0, len(c.opts))
	if withExecutable {
		args = append(args, "ffmpeg")
	}
	for _, opt := range c.opts {
		args = append(args, opt.String())
	}
	return args
}

func (c *Command) Execute() (db.ProgressChan, db.FinishChan, context.CancelFunc, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	args := c.prepareArgs(false)
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		cancelFunc()
		return nil, nil, nil, err
	}
	err = cmd.Start()
	if err != nil {
		cancelFunc()
		return nil, nil, nil, err
	}
	finishChan := make(db.FinishChan, 1)
	progressChan := make(db.ProgressChan, 32)
	go c.completionWatcher(cancelFunc, cmd, stderrReader, finishChan)
	go c.progressWatcher(stderrReader, progressChan)
	return progressChan, finishChan, cancelFunc, nil
}

func (c *Command) ExecuteSync() (*string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	args := c.prepareArgs(false)
	cmd := exec.Command("ffmpeg", args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	outputStr := string(outputBytes)
	return &outputStr, nil
}

func (c *Command) String() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	args := c.prepareArgs(true)
	return strings.Join(args, " ")
}
