package esclation

import (
	"agent-integrations/manager"
	"agent-integrations/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Session ...
type Session struct {
	Driver  string
	Config  json.RawMessage
	History []string
}

// SocketServer ...
func SocketServer(c *gin.Context) {
	wshandler(c.Writer, c.Request)
}

var wsupgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade: ", err)
		return
	}

	params := r.URL.Query()
	var sessionID string
	if len(params["session_id"]) != 0 {
		sessionID = params["session_id"][0]
	}
	log.Println("sessionID", sessionID)

	rdb := redis.NewRedisClient()

	ctx := context.Background()
	body, _ := rdb.Get(ctx, sessionID).Result()

	session := Session{}
	json.Unmarshal([]byte(body), &session)

	chatStarted := false

	messageHandler := func(messageType int, message json.RawMessage) {
		resp, _ := json.Marshal(map[string]interface{}{
			"event": "message",
			"data":  message,
		})
		conn.WriteMessage(messageType, resp)
	}

	startHandler := func(messageType int, message json.RawMessage) {
		log.Println("STARTED")
		chatStarted = true
		resp, _ := json.Marshal(map[string]interface{}{
			"event": "chatStarted",
			"data":  message,
		})
		conn.WriteMessage(messageType, resp)
	}

	agentLeftHandler := func(messageType int, message json.RawMessage) {
		resp, _ := json.Marshal(map[string]interface{}{
			"event": "agentLeft",
			"data":  "AGENT_LEFT",
		})
		conn.WriteMessage(messageType, resp)
	}

	endHandler := func(messageType int, message json.RawMessage) {
		resp, _ := json.Marshal(map[string]interface{}{
			"event": "chatEnded",
			"data":  "CHAT_ENDED",
		})
		conn.WriteMessage(messageType, resp)
	}

	failedHandler := func(messageType int, message json.RawMessage) {
		resp, _ := json.Marshal(map[string]interface{}{
			"event": "chatFailed",
			"data":  "CHAT_FAILED",
		})
		conn.WriteMessage(messageType, resp)
	}

	c := manager.Client{
		Driver:           session.Driver,
		Config:           session.Config,
		History:          session.History,
		MessageHandler:   messageHandler,
		StartHandler:     startHandler,
		AgentLeftHandler: agentLeftHandler,
		EndHandler:       endHandler,
		FailedHandler:    failedHandler,
	}
	integeration := manager.NewClient(c)
	if integeration == nil {
		return
	}
	integeration.Start()

	conn.WriteMessage(1, []byte("PING"))

	conf := map[string]interface{}{}
	json.Unmarshal(session.Config, &conf)

	to, _ := json.Marshal(conf["timeout"])
	var timeout int
	json.Unmarshal(to, &timeout)

	if timeout != 0 {
		u, _ := time.ParseDuration(fmt.Sprintf("%dms", timeout))
		time.AfterFunc(u, func() {
			if !chatStarted {
				log.Println("CHAT_TIMEOUT")
				resp, _ := json.Marshal(map[string]interface{}{
					"event": "chatTimeout",
					"data":  "CHAT_TIMEOUT",
				})
				conn.WriteMessage(1, resp)
				integeration.End()
			}
		})
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		log.Println("msg", msg)

		if string(msg) == "PONG" {
			u, _ := time.ParseDuration(fmt.Sprintf("%ds", 60))
			time.AfterFunc(u, func() {
				conn.WriteMessage(1, []byte("PING"))
			})
		} else {
			integeration.Send(msg)
		}
	}
}
