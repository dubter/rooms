package main

import (
	"app-client/internal/chat"
	"app-client/internal/config"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// not so clever client just for example
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var nickname, password, token string
	var chats map[string]string

	// block with retries if error exceeded
	var client *chat.Client
	for {
		fmt.Println("Nickname:")
		fmt.Scan(&nickname)
		fmt.Println("Password:")
		fmt.Scan(&password)

		client = chat.NewClient(&chat.Config{
			FullAddress: fmt.Sprintf("%s:%s", cfg.Address, cfg.Port),
			Nickname:    nickname,
			Password:    password,
		})

		err = client.Register()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		token, err = client.Login()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// chats -> map[name]id
		chats, err = client.MapChats(token)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		break
	}

	fmt.Println("Доступные чаты:")
	i := 1
	for name := range chats {
		fmt.Printf("%d) %s\n", i, name)
		i++
	}

	fmt.Println("Напиши в консоль название чата для подключения к нему:")
	var chatName string
	fmt.Scan(&chatName)

	// Handling graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigCh
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	for {
		err = client.ConnectToChat(ctx, token, chats[chatName])
		if err != nil {
			fmt.Println(err.Error())

			if errors.Is(err, chat.ErrCtxDone) {
				break
			}
			continue
		}
	}
}
