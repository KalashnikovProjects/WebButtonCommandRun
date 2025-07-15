package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
)

type firstMessageStruct struct {
	command string
}

type messageStruct struct {
	messageType string
	data        string
}

func RunCommandWebsocket(c *websocket.Conn) {
	var (
		mt  int
		msg []byte
		err error
	)
	if mt, msg, err = c.ReadMessage(); err != nil {
		return // fmt.Errorf("error while reading first message %w", err)
	}
	if mt != websocket.TextMessage {
		return // fmt.Errorf("bad first message type, expected TextMessage, not Binary")
	}
	var firstMessageData firstMessageStruct
	if err = json.Unmarshal(msg, &firstMessageData); err != nil {
		return // fmt.Errorf("cant read first message")
	}
	command := firstMessageData.command
	if command == "" {
		return // fmt.Errorf("error empty command")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input, output, err := usecases.RunCommand(ctx, command)
	if err != nil {
		return
	}

	go func() {
		defer cancel()
		for out := range output {
			data, err := json.Marshal(map[string]string{"data": out})
			if err != nil {
				log.Debug(fmt.Errorf("error marshaling message for websocket %w", err))
				continue
			}
			if err = c.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Warn(fmt.Errorf("error writing websocket message %w", err))
				return
			}
		}
	}()

	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			break
		}
		if mt != websocket.TextMessage {
			break
		}
		var messageData messageStruct
		if err = json.Unmarshal(msg, &messageData); err != nil {
			log.Debug(fmt.Errorf("error unmarshal websocket message %w", err))
			continue
		}
		switch messageData.messageType {
		case "data":
			select {
			case input <- messageData.data:
			case <-ctx.Done():
				return
			}
		}
	}
}
