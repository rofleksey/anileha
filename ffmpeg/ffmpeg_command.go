package ffmpeg

import (
	"anileha/db"
	"anileha/util"
	"bufio"
	"context"
	"io"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// TODO: support videos with hardcoded subs

type Command struct {
	mutex sync.Mutex
	opts  []option

	// immutable
	frameRegex     *regexp.Regexp
	numberOfFrames int
}

type CommandSignalEnd struct {
	Err error
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
	OptionPostOutput OptionPriority = 5
	optionOutputFile OptionPriority = 6
)

func (c *Command) AddSingle(key string, optType OptionPriority) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	opt := newOption(key, optType, nil)
	c.opts = append(c.opts, opt)
}

func (c *Command) AddKeyValue(key string, value string, optType OptionPriority) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	opt := newOption(key, optType, &value)
	c.opts = append(c.opts, opt)
}

func newOption(key string, priority OptionPriority, value *string) option {
	return option{key, priority, value}
}

func (o *option) getStrings() []string {
	if o.value != nil {
		return []string{o.key, *o.value}
	}
	return []string{o.key}
}

func NewCommand(inputFile string, numberOfFrames int, outputFile string) *Command {
	frameRegex := regexp.MustCompile("frame=\\s+?(\\d+)")
	command := Command{
		opts:           make([]option, 0, 32),
		frameRegex:     frameRegex,
		numberOfFrames: numberOfFrames,
	}
	command.AddSingle("-hide_banner", OptionBase)
	command.AddSingle("-y", OptionBase)
	command.AddKeyValue("-hwaccel", "auto", OptionBase)
	command.AddKeyValue("-stats_period", "5", OptionBase)
	//command.AddKeyValue("-progress", "pipe:2", OptionBase)
	command.AddKeyValue("-i", inputFile, optionInputFile)
	command.AddSingle(outputFile, optionOutputFile)
	return &command
}

func (c *Command) processWatcher(cmd *exec.Cmd, reader io.ReadCloser, outputChan db.AnyChannel) {
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
			outputChan <- progress
		} else {
			outputChan <- line
		}
	}
	code := cmd.Wait()
	outputChan <- CommandSignalEnd{
		Err: code,
	}
	close(outputChan)
}

//func (c *Command) completionWatcher(cancelFunc context.CancelFunc, cmd *exec.Cmd, reader io.ReadCloser, outputChan db.AnyChannel) {
//	err := cmd.Wait()
//	_ = reader.Close()
//	cancelFunc()
//	finishChan <- err
//	close(finishChan)
//}

func (c *Command) prepareArgs(withExecutable bool) []string {
	sort.SliceStable(c.opts, func(i, j int) bool {
		return c.opts[i].priority < c.opts[j].priority
	})
	args := make([]string, 0, len(c.opts))
	if withExecutable {
		args = append(args, "ffmpeg")
	}
	for _, opt := range c.opts {
		args = append(args, opt.getStrings()...)
	}
	return args
}

func (c *Command) Execute() (db.AnyChannel, context.CancelFunc, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	args := c.prepareArgs(false)
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		cancelFunc()
		return nil, nil, err
	}
	err = cmd.Start()
	if err != nil {
		cancelFunc()
		return nil, nil, err
	}
	outputChan := make(db.AnyChannel, 32)
	go c.processWatcher(cmd, stderrReader, outputChan)
	return outputChan, cancelFunc, nil
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
