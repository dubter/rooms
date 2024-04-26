package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
)

var ErrCtxDone = errors.New("context Done, closing WebSocket connection")

type Client struct {
	fullAddress string
	nickname    string
	password    string
}

type Config struct {
	FullAddress string
	Nickname    string
	Password    string
}

func NewClient(cfg *Config) *Client {
	return &Client{
		fullAddress: cfg.FullAddress,
		nickname:    cfg.Nickname,
		password:    cfg.Password,
	}
}

func (c *Client) Register() error {
	register := &AuthReq{
		Nickname: c.nickname,
		Password: c.password,
	}

	payload, err := json.Marshal(register)
	if err != nil {
		return fmt.Errorf("can not marshal AuthReq: %w", err)
	}

	resp, err := http.Post("http://"+c.fullAddress+"/user/register", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error http.Post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Login() (string, error) {
	register := &AuthReq{
		Nickname: c.nickname,
		Password: c.password,
	}

	payload, err := json.Marshal(register)
	if err != nil {
		return "", fmt.Errorf("can not marshal AuthReq: %w", err)
	}

	resp, err := http.Post("http://"+c.fullAddress+"/user/login", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("error http.Post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	payload, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error io.ReadAll: %w", err)
	}

	var tokens tokenResponse
	err = json.Unmarshal(payload, &tokens)
	if err != nil {
		return "", fmt.Errorf("can not unmarshal payload: %w", err)
	}

	return tokens.AccessToken, nil
}

func (c *Client) MapChats(token string) (map[string]string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://"+c.fullAddress+"/chat/rooms", nil)
	if err != nil {
		return nil, fmt.Errorf("error http.NewRequest: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error set connection with chat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error io.ReadAll: %w", err)
	}

	var rooms []RoomResp
	err = json.Unmarshal(payload, &rooms)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal payload: %w", err)
	}

	chats := make(map[string]string) // name -> id

	for i := range rooms {
		room := &rooms[i]
		chats[room.Name] = room.ID
	}

	return chats, nil
}

func (c *Client) ConnectToChat(ctxParent context.Context, token, chatID string) error {
	u := url.URL{Scheme: "ws", Host: c.fullAddress, Path: "/chat/rooms/" + chatID}
	fmt.Printf("connecting to %s", u.String())

	headers := make(http.Header)
	headers.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return fmt.Errorf("eror websocket.DefaultDialer: %w", err)
	}
	defer conn.Close()

	eg, ctx := errgroup.WithContext(ctxParent)

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ErrCtxDone
			default:
				if err = ReadConn(conn); err != nil {
					return err
				}
			}
		}
	})

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ErrCtxDone
			default:
				if err = WriteConn(conn); err != nil {
					return err
				}
			}
		}
	})

	return eg.Wait()
}

func ReadConn(conn *websocket.Conn) error {
	_, payload, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("ошибка при чтении из WebSocket: %w", err)
	}

	var messages []Message
	err = json.Unmarshal(payload, &messages)
	if err != nil {
		var msg Message
		err = json.Unmarshal(payload, &msg)
		if err != nil {
			return nil
		}

		fmt.Printf("(%s) %s: %s\n", msg.TimeCreated, msg.Username, msg.Content)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[j].TimeCreated.After(messages[i].TimeCreated)
	})

	for i := range messages {
		msg := &messages[i]
		fmt.Printf("(%s) %s: %s\n", msg.TimeCreated, msg.Username, msg.Content)
	}

	return nil
}

func WriteConn(conn *websocket.Conn) error {
	reader := bufio.NewReader(os.Stdin)
	message, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("ошибка при чтении ввода stdin: %w", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return fmt.Errorf("ошибка при записи сообщение в WebSocket: %w", err)
	}

	return nil
}
