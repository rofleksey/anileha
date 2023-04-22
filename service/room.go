package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util/ws"
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"sync"
	"time"
)

type RoomService struct {
	config *config.Config
	log    *zap.Logger
	mutex  sync.Mutex
	rooms  map[string]*room
}

func NewRoomService(
	config *config.Config,
	log *zap.Logger) *RoomService {
	return &RoomService{
		config: config,
		log:    log,
		rooms:  make(map[string]*room),
	}
}

type room struct {
	mutex    sync.Mutex
	id       string
	state    RoomState
	watchers map[uint]*watcher
}

func (r *room) broadcastExcept(id uint, req any) {
	for _, w := range r.watchers {
		if w.state.Id == id {
			continue
		}
		w.client.Send(req)
	}
}

type RoomState struct {
	EpisodeId   *uint   `json:"episodeId"`
	Timestamp   float64 `json:"timestamp"`
	Playing     bool    `json:"playing"`
	InitiatorId uint    `json:"initiatorId"`
}

type watcher struct {
	client *ws.Client
	state  WatcherState
}

type WatcherState struct {
	Id        uint    `json:"id"`
	Name      string  `json:"name"`
	Thumb     string  `json:"thumb"`
	Timestamp float64 `json:"timestamp"`
	Progress  float64 `json:"progress"`
	Status    string  `json:"status"`
}

type WatcherStatePartial struct {
	Timestamp float64 `json:"timestamp"`
	Progress  float64 `json:"progress"`
	Status    string  `json:"status"`
}

type IdMessage struct {
	Id uint `json:"id"`
}

type FullState struct {
	Room     RoomState      `json:"room"`
	Watchers []WatcherState `json:"watchers"`
}

type MessageHeader struct {
	Type string `json:"type"`
}

type MessageStructure[T any] struct {
	Type    string `json:"type"`
	Message T      `json:"message"`
}

func (s *RoomService) handleRoomStateRequest(watcher *watcher, userRoom *room, bytes []byte) {
	var req MessageStructure[RoomState]
	err := json.Unmarshal(bytes, &req)
	if err != nil {
		s.log.Warn("failed to unmarshal room state request", zap.Error(err))
		return
	}

	s.log.Info("room state", zap.Any("state", req.Message))

	userRoom.mutex.Lock()
	defer userRoom.mutex.Unlock()

	if req.Message.EpisodeId != nil {
		userRoom.state.EpisodeId = req.Message.EpisodeId
	} else {
		req.Message.EpisodeId = userRoom.state.EpisodeId
	}

	if req.Message.Timestamp < 0 {
		req.Message.Timestamp = userRoom.state.Timestamp
	}

	userRoom.state = req.Message
	req.Message.InitiatorId = watcher.state.Id

	userRoom.broadcastExcept(watcher.state.Id, req)
}

func (s *RoomService) handleUserStateRequest(watcher *watcher, userRoom *room, bytes []byte) {
	var req MessageStructure[WatcherStatePartial]
	err := json.Unmarshal(bytes, &req)
	if err != nil {
		s.log.Warn("failed to unmarshal user state request", zap.Error(err))
		return
	}

	userRoom.mutex.Lock()
	defer userRoom.mutex.Unlock()

	watcher.state.Status = req.Message.Status
	watcher.state.Progress = req.Message.Progress
	watcher.state.Timestamp = req.Message.Timestamp

	userRoom.broadcastExcept(watcher.state.Id, MessageStructure[WatcherState]{
		Type:    "user-state",
		Message: watcher.state,
	})
}

func (s *RoomService) handleMessage(watcher *watcher, userRoom *room, bytes []byte) {
	var req MessageHeader
	err := json.Unmarshal(bytes, &req)
	if err != nil {
		s.log.Warn("failed to unmarshal request", zap.Error(err))
		return
	}

	switch req.Type {
	case "room-state":
		s.handleRoomStateRequest(watcher, userRoom, bytes)
	case "user-state":
		s.handleUserStateRequest(watcher, userRoom, bytes)
	case "user-disconnect":
		watcher.client.Close(false)
		s.handleDisconnect(watcher, userRoom)
	default:
		s.log.Warn("invalid message type", zap.String("type", req.Type))
	}
}

func (s *RoomService) roomCleanupWorker(userRoom *room) {
	time.Sleep(1 * time.Minute)

	userRoom.mutex.Lock()
	defer userRoom.mutex.Unlock()

	if len(userRoom.watchers) == 0 {
		s.mutex.Lock()
		delete(s.rooms, userRoom.id)
		s.mutex.Unlock()
	}
}

func (s *RoomService) handleDisconnect(watcher *watcher, userRoom *room) {
	userRoom.mutex.Lock()
	defer userRoom.mutex.Unlock()

	delete(userRoom.watchers, watcher.state.Id)

	userRoom.broadcastExcept(watcher.state.Id, MessageStructure[IdMessage]{
		Type: "user-disconnect",
		Message: IdMessage{
			Id: watcher.state.Id,
		},
	})

	if len(userRoom.watchers) == 0 {
		go s.roomCleanupWorker(userRoom)
	}
}

func (s *RoomService) GetAll() []RoomState {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	roomStates := make([]RoomState, 0, len(s.rooms))
	for _, r := range s.rooms {
		r.mutex.Lock()
		if len(r.watchers) > 0 {
			roomStates = append(roomStates, r.state)
		}
		r.mutex.Unlock()
	}

	return roomStates
}

func (s *RoomService) HandleConnection(conn *websocket.Conn, user *db.User, roomId string) {
	s.mutex.Lock()
	userRoom, roomExists := s.rooms[roomId]
	if !roomExists {
		userRoom = &room{
			id:       roomId,
			watchers: make(map[uint]*watcher),
		}
		s.rooms[roomId] = userRoom
	}

	userRoom.mutex.Lock()
	defer userRoom.mutex.Unlock()

	s.mutex.Unlock()

	userRoom.state.Playing = false

	var curWatcher *watcher

	receiveListener := func(client *ws.Client, bytes []byte) {
		s.handleMessage(curWatcher, userRoom, bytes)
	}

	disconnectListener := func(client *ws.Client) {
		s.handleDisconnect(curWatcher, userRoom)
	}

	client := ws.NewClient(user.ID, conn, receiveListener, disconnectListener, s.config, s.log)
	curWatcher = &watcher{
		client: client,
		state: WatcherState{
			Id:     user.ID,
			Name:   user.Name,
			Thumb:  user.Thumb.Url,
			Status: "connected",
		},
	}

	userRoom.broadcastExcept(user.ID, MessageStructure[WatcherState]{
		Type:    "user-connect",
		Message: curWatcher.state,
	})

	existingWatcher, watcherExists := userRoom.watchers[user.ID]
	if watcherExists {
		existingWatcher.client.Close(false)
	}
	userRoom.watchers[user.ID] = curWatcher

	watcherStates := make([]WatcherState, 0, len(userRoom.watchers))
	for _, w := range userRoom.watchers {
		watcherStates = append(watcherStates, w.state)
	}

	client.Send(MessageStructure[FullState]{
		Type: "full-state",
		Message: FullState{
			Room:     userRoom.state,
			Watchers: watcherStates,
		},
	})

	client.Start()
}

var RoomExport = fx.Options(fx.Provide(NewRoomService))
