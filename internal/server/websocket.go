package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/json_storage"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
	"strconv"
	"sync"
	"time"
)

type inputMessageStruct struct {
	MessageType string                  `json:"message-type"`
	Data        string                  `json:"data"`
	Options     entities.CommandOptions `json:"options"`
}

type outMessageStruct struct {
	MessageType string `json:"message-type"`
	Data        string `json:"data"`
}

func RunCommandWebsocket(c *websocket.Conn) {
	defer c.Close()
	var (
		mt  int
		msg []byte
		err error
	)
	commandId, err := strconv.Atoi(c.Params("id"))
	if err != nil || commandId < 0 {
		data := websocket.FormatCloseMessage(1003, "bad command id")
		if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
			log.Warn("Error writing close message", err)
		}
		return
	}
	commandData, err := json_storage.GetCommand(uint(commandId))
	if err != nil {
		data := websocket.FormatCloseMessage(1003, "command not found")
		if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
			log.Warn("Error writing close message", err)
		}
		return
	}
	if commandData.Command == "" {
		data := websocket.FormatCloseMessage(1002, "empty command")
		if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
			log.Warn("Error writing close message: ", err)
		}
		return
	}

	if mt, msg, err = c.ReadMessage(); err != nil {
		data := websocket.FormatCloseMessage(1011, "error reading message")
		_ = c.WriteMessage(websocket.CloseMessage, data)
		return
	}
	if mt != websocket.TextMessage {
		data := websocket.FormatCloseMessage(1003, "first message must be options")
		if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
			log.Warn("Error writing close message: ", err)
		}
		return
	}
	inputData := &inputMessageStruct{}
	err = json.Unmarshal(msg, inputData)
	if err != nil {
		data := websocket.FormatCloseMessage(1003, "bad input json")
		if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
			log.Warn("Error writing close message: ", err)
		}
		return
	}
	if inputData.MessageType != "options" {
		data := websocket.FormatCloseMessage(1003, "first message must be options")
		if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
			log.Warn("Error writing close message: ", err)
		}
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Starting command:", commandData.Command)
	runningCommand, err := usecases.RunCommand(ctx, commandData.Command, inputData.Options)
	if err != nil {
		log.Warn("Error while stating command: ", err)
		data := websocket.FormatCloseMessage(1011, "unexpected error while stating command")
		if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
			log.Warn("Error writing close message: ", err)
		}
		return
	}

	outMutex := &sync.Mutex{}
	websocketWriteMutex := &sync.Mutex{}
	outBuffer := ""
	outBufferEOF := false

	// Get output
	go func() {
		for out := range runningCommand.Output {
			outMutex.Lock()
			outBuffer += out
			outMutex.Unlock()
		}
		outBufferEOF = true
	}()

	// Writer in interval
	go func() {
		defer func() {
			data := websocket.FormatCloseMessage(1000, "command run finished")
			websocketWriteMutex.Lock()
			_ = c.WriteMessage(websocket.CloseMessage, data)
			websocketWriteMutex.Unlock()
			cancel()
		}()

		ticker := time.NewTicker(config.Config.WebsocketWriteInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if outBuffer == "" {
					if outBufferEOF {
						return
					}
					continue
				}
				outMutex.Lock()
				data, err := json.Marshal(outMessageStruct{"data", outBuffer})
				if err != nil {
					log.Debug(fmt.Errorf("error marshaling message for websocket %w", err))
					outMutex.Unlock()
					continue
				}
				outBuffer = ""
				outMutex.Unlock()

				websocketWriteMutex.Lock()
				err = c.WriteMessage(websocket.TextMessage, data)
				websocketWriteMutex.Unlock()
				if err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Input loop
	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			data := websocket.FormatCloseMessage(1011, "error reading message")
			websocketWriteMutex.Lock()
			_ = c.WriteMessage(websocket.CloseMessage, data)
			websocketWriteMutex.Unlock()
			return
		}
		if mt == websocket.CloseMessage {
			cancel()
			return
		}
		if mt != websocket.TextMessage {
			data := websocket.FormatCloseMessage(1003, "expected TextMessage, not BinaryData")
			websocketWriteMutex.Lock()
			if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
				log.Warn("Error writing close message: ", err)
			}
			websocketWriteMutex.Unlock()
			return
		}
		inputData := &inputMessageStruct{}
		err = json.Unmarshal(msg, inputData)
		if err != nil {
			data := websocket.FormatCloseMessage(1003, "bad input json")
			websocketWriteMutex.Lock()
			if err = c.WriteMessage(websocket.CloseMessage, data); err != nil {
				log.Warn("Error writing close message: ", err)
			}
			websocketWriteMutex.Unlock()
			return
		}
		switch inputData.MessageType {
		case "terminal-input":
			select {
			case runningCommand.Input <- inputData.Data:
			case <-ctx.Done():
				return
			}
		}
	}
}
