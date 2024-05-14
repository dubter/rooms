package chat

import (
	"app-websocket/internal/domain"
	common "app-websocket/internal/ports/http"
	"app-websocket/internal/ports/ws"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"io"
	"log/slog"
	"net/http"
)

type ServiceChatCache interface {
	GetLastMessagesFromRoom(ctx context.Context, roomID string) ([]domain.Message, error)
	GetRoomClients(ctx context.Context, roomID string) ([]domain.User, error)
}

type ServiceRoomsProvider interface {
	GetAllRooms(ctx context.Context) ([]domain.Room, error)
	GetRoom(ctx context.Context, roomID string) (*domain.Room, error)
	CreateRoom(ctx context.Context, name string) (*domain.Room, error)
}

type ServiceChatPusher interface {
	PushMessage(ctx context.Context, msg *domain.Message) error
	Subscribe(ctx context.Context, client *ws.Client) error
	Unsubscribe(ctx context.Context, client *ws.Client) error
}

type Handler struct {
	logger        *slog.Logger
	chatCache     ServiceChatCache
	chatPusher    ServiceChatPusher
	roomsProvider ServiceRoomsProvider
}

func NewHandler(logger *slog.Logger, chatCache ServiceChatCache, chatPusher ServiceChatPusher, roomsProvider ServiceRoomsProvider) *Handler {
	return &Handler{
		logger:        logger,
		chatCache:     chatCache,
		chatPusher:    chatPusher,
		roomsProvider: roomsProvider,
	}
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		common.ProcessError(w, "can not read request body", http.StatusBadRequest)
		return
	}

	var req CreateRoomReq
	err = json.Unmarshal(buf, &req)
	if err != nil {
		common.ProcessError(w, "can not unmarshal request body", http.StatusBadRequest)
		return
	}

	room, err := h.roomsProvider.CreateRoom(r.Context(), req.Name)
	if err != nil {
		h.logger.Error("failed to create room", slog.String("error", err.Error()))
		common.ProcessError(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	roomsResp := RoomRes{
		ID:          room.ID,
		Name:        room.Name,
		TimeCreated: room.TimeCreated,
	}

	payload, err := json.Marshal(roomsResp)
	if err != nil {
		h.logger.Error("can not marshal response body", slog.String("error", err.Error()))
		common.ProcessError(w, "can not marshal response body", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	if len(roomID) == 0 {
		common.ProcessError(w, "'id' is required param", http.StatusBadRequest)
		return
	}

	_, err := h.roomsProvider.GetRoom(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			common.ProcessError(w, domain.ErrRoomNotFound.Error(), http.StatusBadRequest)
			return
		}

		h.logger.Error("failed to get room", slog.String("error", err.Error()))
		common.ProcessError(w, "failed to get room", http.StatusInternalServerError)
		return
	}

	userID := r.Header.Get("user_id")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
	}

	username := r.Header.Get("nickname")
	if username == "" {
		username = r.URL.Query().Get("nickname")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection to web socket", slog.String("error", err.Error()))
		common.ProcessError(w, "failed to upgrade connection to web socket", http.StatusInternalServerError)
		return
	}

	messages, err := h.chatCache.GetLastMessagesFromRoom(context.Background(), roomID)
	if err != nil {
		h.logger.Error("can not get last messages from room",
			slog.String("username", username),
			slog.String("RoomID", roomID),
			slog.String("error", err.Error()))
	}

	messagesResp := make([]ws.Message, 0)
	for i := range messages {
		messagesResp = append(messagesResp, ws.Message{
			Content:     messages[i].Content,
			TimeCreated: messages[i].TimeCreated,
			Username:    messages[i].Nickname,
			UserID:      messages[i].UserID,
			RoomID:      roomID,
		})
	}

	err = conn.WriteJSON(messagesResp)
	if err != nil {
		h.logger.Error("can not send message to client",
			slog.String("username", username),
			slog.String("RoomID", roomID),
			slog.String("error", err.Error()))
	}

	cl := &ws.Client{
		Conn:    conn,
		Message: make(chan *ws.Message, 10),
		Logger:  h.logger,
		User: &domain.User{
			ID:       userID,
			Nickname: username,
		},
		RoomID: roomID,
		Pusher: h.chatPusher,
	}

	err = h.chatPusher.Subscribe(r.Context(), cl)
	if err != nil {
		h.logger.Error("failed to subscribe:", slog.String("error", err.Error()))
	}

	go cl.WriteMessage()
	cl.ReadMessage(r.Context())
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomsProvider.GetAllRooms(r.Context())
	if err != nil {
		h.logger.Error("failed to get rooms", slog.String("error", err.Error()))
		common.ProcessError(w, "failed to get rooms", http.StatusInternalServerError)
		return
	}

	var roomsResps []RoomRes
	for _, room := range rooms {
		roomsResp := RoomRes{
			ID:          room.ID,
			Name:        room.Name,
			TimeCreated: room.TimeCreated,
		}

		roomsResps = append(roomsResps, roomsResp)
	}

	payload, err := json.Marshal(roomsResps)
	if err != nil {
		h.logger.Error("can not marshal response body", slog.String("error", err.Error()))
		common.ProcessError(w, "can not marshal response body", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(payload)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetClients(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	if len(roomID) == 0 {
		common.ProcessError(w, "'id' is required param", http.StatusBadRequest)
		return
	}

	_, err := h.roomsProvider.GetRoom(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			common.ProcessError(w, domain.ErrRoomNotFound.Error(), http.StatusBadRequest)
			return
		}

		h.logger.Error("failed to get room", slog.String("error", err.Error()))
		common.ProcessError(w, "failed to get room", http.StatusInternalServerError)
		return
	}

	users, err := h.chatCache.GetRoomClients(r.Context(), roomID)
	if err != nil {
		h.logger.Error("failed to get room clients", slog.String("error", err.Error()))
		common.ProcessError(w, "failed to get room clients", http.StatusInternalServerError)
		return
	}

	clients := make([]ClientRes, 0)
	for _, c := range users {
		clients = append(clients, ClientRes{
			Username: c.Nickname,
		})
	}

	payload, err := json.Marshal(clients)
	if err != nil {
		h.logger.Error("can not marshal response body", slog.String("error", err.Error()))
		common.ProcessError(w, "can not marshal response body", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
