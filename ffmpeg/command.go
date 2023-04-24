package ffmpeg

import (
	"anileha/util"
	"bufio"
	"bytes"
	"context"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var timeRegex = regexp.MustCompile("time=(\\d+):(\\d+):(\\d+).(\\d+)")

type Command struct {
	mutex    sync.Mutex
	cmd      string
	args     string
	vars     map[string][]string
	logsPath *string

	// immutable
	videoDurationSec int
}

type CommandSignalEnd struct {
	Err error
}

func (c *Command) AddVar(key string, value ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.vars[key] = value
}

func (c *Command) WriteLogsTo(logsPath string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.logsPath = &logsPath
}

func NewCommand(cmd string, args string, videoDurationSec int) *Command {
	command := Command{
		cmd:              cmd,
		args:             args,
		vars:             make(map[string][]string),
		videoDurationSec: videoDurationSec,
	}

	command.vars["BASE"] = []string{"-hide_banner", "-y", "-hwaccel", "auto", "-stats_period", "3"}

	return &command
}

func (c *Command) parseTime(line string) uint64 {
	// frame=  524 fps= 79 q=-1.0 Lsize=    8014kB time=00:00:22.66 bitrate=2896.6kbits/s speed=3.43x
	// need to parse time here
	matchResult := timeRegex.FindStringSubmatch(line)
	if matchResult != nil {
		hours, _ := strconv.Atoi(matchResult[1])
		minutes, _ := strconv.Atoi(matchResult[2])
		seconds, _ := strconv.Atoi(matchResult[3])
		return uint64(hours*60*60 + minutes*60 + seconds)
	}
	return 0
}

func (c *Command) logsWriter(logsChan chan string, externalLog *zap.Logger) {
	if c.logsPath == nil {
		externalLog.Warn("file logs are turned off for command")
		for range logsChan {
		}
		return
	}
	var writer *bufio.Writer
	file, err := os.Create(*c.logsPath)
	defer func() {
		file.Sync()
		file.Close()
	}()
	if err != nil {
		externalLog.Warn("failed to open file", zap.String("file", *c.logsPath), zap.Error(err))
		for range logsChan {
		}
		return
	}
	writer = bufio.NewWriter(file)
	for line := range logsChan {
		writer.WriteString(line + "\n")
		writer.Flush()
	}
	externalLog.Warn("closing logs file", zap.String("file", *c.logsPath))
}

func (c *Command) processWatcher(cmd *exec.Cmd, reader io.ReadCloser, outputChan chan any, externalLog *zap.Logger) {
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
	if c.videoDurationSec != 0 {
		etaCalculator = util.NewEtaCalculator(0, float64(c.videoDurationSec))
	} else {
		etaCalculator = util.NewUndefinedEtaCalculator()
	}
	logsChan := make(chan string, 32)
	go c.logsWriter(logsChan, externalLog)
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

func (c *Command) interpolateArgs() []string {
	result := make([]string, 0, len(c.args))
	argsSplit := strings.Split(c.args, " ")

	for _, arg := range argsSplit {
		if strings.HasPrefix(arg, "$") {
			varName := strings.TrimPrefix(arg, "$")
			varValue, varExists := c.vars[varName]
			if !varExists {
				continue
			}
			result = append(result, varValue...)
		} else {
			result = append(result, arg)
		}
	}

	return result
}

func (c *Command) Execute(externalLog *zap.Logger) (chan any, context.CancelFunc, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	resultArgs := c.interpolateArgs()
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, c.cmd, resultArgs...)

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

	outputChan := make(chan any, 32)

	go c.processWatcher(cmd, stdoutReader, outputChan, externalLog)

	return outputChan, cancelFunc, nil
}

func (c *Command) ExecuteSync() ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	resultArgs := c.interpolateArgs()
	cmd := exec.Command(c.cmd, resultArgs...)
	outputBytes, err := cmd.CombinedOutput()

	return outputBytes, err
}

func (c *Command) String() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	args := c.interpolateArgs()

	return c.cmd + " " + strings.Join(args, " ")
}
