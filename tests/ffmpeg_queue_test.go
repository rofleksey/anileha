package tests

import (
	"anileha/db"
	"anileha/ffmpeg"
	"anileha/util"
	"context"
	"reflect"
	"testing"
	"time"
)

type testItem struct{}

func (t *testItem) Execute() (db.AnyChannel, context.CancelFunc, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	channel := make(db.AnyChannel)
	go func() {
		timer := time.NewTimer(1 * time.Second)
		select {
		case <-timer.C:
			channel <- "test"
			close(channel)
		case <-ctx.Done():
			channel <- util.ErrCancelled
			close(channel)
		}
	}()
	return channel, cancelFunc, nil
}

func TestQueueSimple(t *testing.T) {
	outputChan := make(chan ffmpeg.OutputMessage)
	queue, err := ffmpeg.NewQueue(1, outputChan)
	if err != nil {
		t.Fatal(err)
	}
	queue.Start()
	go func() {
		for i := 0; i < 3; i++ {
			queue.Enqueue(uint(i), &testItem{})
		}
	}()
	output := make([]interface{}, 0, 10)
	for i := 0; i < 6; i++ {
		output = append(output, <-outputChan)
	}
	expected := []interface{}{
		ffmpeg.OutputMessage{
			ID:  0,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  0,
			Msg: "test",
		},
		ffmpeg.OutputMessage{
			ID:  1,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  1,
			Msg: "test",
		},
		ffmpeg.OutputMessage{
			ID:  2,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  2,
			Msg: "test",
		},
	}
	if !reflect.DeepEqual(expected, output) {
		t.Fatal("Invalid queue response")
	}
}

func TestQueueInterruptFuture(t *testing.T) {
	outputChan := make(chan ffmpeg.OutputMessage)
	queue, err := ffmpeg.NewQueue(1, outputChan)
	if err != nil {
		t.Fatal(err)
	}
	queue.Start()
	go func() {
		for i := 0; i < 3; i++ {
			queue.Enqueue(uint(i), &testItem{})
		}
		queue.Cancel(1)
	}()
	output := make([]interface{}, 0, 10)
	for i := 0; i < 5; i++ {
		output = append(output, <-outputChan)
	}
	expected := []interface{}{
		ffmpeg.OutputMessage{
			ID:  0,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  0,
			Msg: "test",
		},
		ffmpeg.OutputMessage{
			ID:  1,
			Msg: util.ErrCancelled,
		},
		ffmpeg.OutputMessage{
			ID:  2,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  2,
			Msg: "test",
		},
	}
	if !reflect.DeepEqual(expected, output) {
		t.Fatal("Invalid queue response", output)
	}
}

func TestQueueInterruptCurrent(t *testing.T) {
	outputChan := make(chan ffmpeg.OutputMessage)
	queue, err := ffmpeg.NewQueue(1, outputChan)
	if err != nil {
		t.Fatal(err)
	}
	queue.Start()
	go func() {
		for i := 0; i < 3; i++ {
			queue.Enqueue(uint(i), &testItem{})
		}
		queue.Cancel(0)
	}()
	output := make([]interface{}, 0, 10)
	for i := 0; i < 6; i++ {
		output = append(output, <-outputChan)
	}
	expected := []interface{}{
		ffmpeg.OutputMessage{
			ID:  0,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  0,
			Msg: util.ErrCancelled,
		},
		ffmpeg.OutputMessage{
			ID:  1,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  1,
			Msg: "test",
		},
		ffmpeg.OutputMessage{
			ID:  2,
			Msg: ffmpeg.QueueSignalStarted{},
		},
		ffmpeg.OutputMessage{
			ID:  2,
			Msg: "test",
		},
	}
	if !reflect.DeepEqual(expected, output) {
		t.Fatal("Invalid queue response", output)
	}
}
