package ffmpeg

import (
	"anileha/db"
	"anileha/util"
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Command struct {
	mutex    sync.Mutex
	opts     []option
	logsPath *string

	// immutable
	timeRegex       *regexp.Regexp
	videoDuratioSec uint64
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

func (c *Command) WriteLogsTo(logsPath string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.logsPath = &logsPath
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

func NewCommand(inputFile string, videoDurationSec uint64, outputFile string) *Command {
	// frame=  524 fps= 79 q=-1.0 Lsize=    8014kB time=00:00:22.66 bitrate=2896.6kbits/s speed=3.43x
	// need to parse time here
	timeRegex := regexp.MustCompile("time=(\\d+):(\\d+):(\\d+).(\\d+)")
	command := Command{
		opts:            make([]option, 0, 32),
		timeRegex:       timeRegex,
		videoDuratioSec: videoDurationSec,
	}
	command.AddSingle("-hide_banner", OptionBase)
	command.AddSingle("-y", OptionBase)
	command.AddKeyValue("-hwaccel", "auto", OptionBase)
	command.AddKeyValue("-stats_period", "2", OptionBase)
	//command.AddKeyValue("-progress", "pipe:2", OptionBase)
	command.AddKeyValue("-i", inputFile, optionInputFile)
	command.AddSingle(outputFile, optionOutputFile)
	return &command
}

func (c *Command) parseTime(line string) uint64 {
	matchResult := c.timeRegex.FindStringSubmatch(line)
	if matchResult != nil {
		hours, _ := strconv.Atoi(matchResult[1])
		minutes, _ := strconv.Atoi(matchResult[2])
		seconds, _ := strconv.Atoi(matchResult[3])
		return uint64(hours*60*60 + minutes*60 + seconds)
	}
	return 0
}

func (c *Command) logsWriter(logsChan chan string, logsPath *string) {
	if logsPath == nil {
		for range logsChan {
		}
		return
	}
	var writer *bufio.Writer
	file, err := os.Create(*c.logsPath)
	if err != nil {
		for range logsChan {
		}
		return
	}
	writer = bufio.NewWriter(file)
	for line := range logsChan {
		writer.WriteString(line + "\n")
		writer.Flush()
	}
	file.Sync()
}

func (c *Command) processWatcher(cmd *exec.Cmd, reader io.ReadCloser, outputChan db.AnyChannel) {
	scanner := bufio.NewScanner(reader)
	// to properly handle carriage return
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if i := bytes.IndexByte(data, '\r'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	var etaCalculator *util.EtaCalculator
	if c.videoDuratioSec != 0 {
		etaCalculator = util.NewEtaCalculator(0, float64(c.videoDuratioSec))
	} else {
		etaCalculator = util.NewUndefinedEtaCalculator()
	}
	logsChan := make(chan string, 32)
	go c.logsWriter(logsChan, c.logsPath)
	etaCalculator.Start()
	for scanner.Scan() {
		line := scanner.Text()
		logsChan <- line
		time := c.parseTime(line)
		if time != 0 {
			etaCalculator.Update(float64(time))
			progress := etaCalculator.GetProgress()
			outputChan <- progress
		}
	}
	code := cmd.Wait()
	outputChan <- CommandSignalEnd{
		Err: code,
	}
	close(logsChan)
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
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		cancelFunc()
		return nil, nil, err
	}
	// merge stderr into stdout
	cmd.Stderr = cmd.Stdout
	err = cmd.Start()
	if err != nil {
		cancelFunc()
		return nil, nil, err
	}
	outputChan := make(db.AnyChannel, 32)
	go c.processWatcher(cmd, stdoutReader, outputChan)
	return outputChan, cancelFunc, nil
}

func (c *Command) ExecuteSync() (*string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	args := c.prepareArgs(false)
	cmd := exec.Command("ffmpeg", args...)
	outputBytes, err := cmd.CombinedOutput()
	outputStr := string(outputBytes)
	return &outputStr, err
}

func (c *Command) String() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	args := c.prepareArgs(true)
	return strings.Join(args, " ")
}
