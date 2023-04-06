package ffmpeg

import (
	"anileha/util"
	"context"
	"go.uber.org/zap"
)

type Queuable interface {
	Execute(externalLog *zap.Logger) (chan any, context.CancelFunc, error)
}

type QueueSignalStarted struct{}

type OutputMessage struct {
	ID  uint
	Msg interface{}
}

type enqueueMessage struct {
	queueItem queueItem
}

type cancelMessage struct {
	ID uint
}

type queueItem struct {
	ID        uint
	Command   Queuable
	CloseChan chan interface{}
}

type Queue struct {
	inputChan          chan interface{}
	workerChan         chan queueItem
	workerFeedBackChan chan uint
	outputChan         chan OutputMessage
	log                *zap.Logger
}

func NewQueue(outputChan chan OutputMessage, log *zap.Logger) (*Queue, error) {
	return &Queue{
		inputChan:          make(chan interface{}),
		workerChan:         make(chan queueItem, 1024),
		workerFeedBackChan: make(chan uint),
		outputChan:         outputChan,
		log:                log,
	}, nil
}

func (q *Queue) Enqueue(id uint, entry Queuable) {
	item := queueItem{
		ID:        id,
		Command:   entry,
		CloseChan: make(chan interface{}, 1),
	}
	q.inputChan <- enqueueMessage{item}
}

func (q *Queue) Cancel(id uint) {
	q.inputChan <- cancelMessage{id}
}

func (q *Queue) inputWorker() {
	items := make(map[uint]queueItem, 32)
	for {
		select {
		case id := <-q.workerFeedBackChan:
			delete(items, id)
		case msg := <-q.inputChan:
			switch castedMsg := msg.(type) {
			case enqueueMessage:
				items[castedMsg.queueItem.ID] = castedMsg.queueItem
				q.workerChan <- castedMsg.queueItem
			case cancelMessage:
				if item, ok := items[castedMsg.ID]; ok {
					close(item.CloseChan)
				}
			}
		}
	}
}

func (q *Queue) processItem(cur *queueItem) {
	defer func() {
		q.workerFeedBackChan <- cur.ID
	}()
	select {
	case <-cur.CloseChan:
		return
	default:
	}
	q.outputChan <- OutputMessage{
		ID:  cur.ID,
		Msg: QueueSignalStarted{},
	}
	cmdChan, cancelFunc, err := cur.Command.Execute(q.log)
	if err != nil {
		cancelFunc()
		q.outputChan <- OutputMessage{
			ID:  cur.ID,
			Msg: err,
		}
		return
	}
	for {
		select {
		case cmdMsg, ok := <-cmdChan:
			if !ok {
				return
			}
			q.outputChan <- OutputMessage{
				ID:  cur.ID,
				Msg: cmdMsg,
			}
		case <-cur.CloseChan:
			cancelFunc()
			for cmdMsg := range cmdChan {
				switch cmdMsg.(type) {
				case CommandSignalEnd:
					q.outputChan <- OutputMessage{
						ID: cur.ID,
						Msg: CommandSignalEnd{
							Err: util.ErrCancelled,
						},
					}
				default:
					q.outputChan <- OutputMessage{
						ID:  cur.ID,
						Msg: cmdMsg,
					}
				}
			}
			return
		}
	}
}

func (q *Queue) processWorker() {
	for {
		cur := <-q.workerChan
		q.processItem(&cur)
	}
}

func (q *Queue) Start() {
	go q.inputWorker()
	go q.processWorker()
}
