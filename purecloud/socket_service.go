package purecloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// SocketService ...
type SocketService struct {
	Config           Config
	History          []string
	ConversationID   string
	UserID           string
	AgentID          string
	JWT              string
	Started          bool
	MessageHandler   func(int, json.RawMessage)
	StartHandler     func(int, json.RawMessage)
	AgentLeftHandler func(int, json.RawMessage)
	EndHandler       func(int, json.RawMessage)
	FailedHandler    func(int, json.RawMessage)
	SocketConn       *websocket.Conn
}

// Start ...
func (socket *SocketService) Start() {
	s := strings.Split(socket.Config.Name, " ")
	var firstName, lastName string
	if len(s) > 0 {
		firstName = s[0]
	}
	if len(s) > 1 {
		lastName = s[1]
	}

	body, _ := json.Marshal(map[string]interface{}{
		"organizationId": socket.Config.OrganizationID,
		"deploymentId":   socket.Config.DeploymentID,
		"routingTarget": map[string]interface{}{
			"targetType":    "QUEUE",
			"targetAddress": socket.Config.QueueName,
		},
		"memberInfo": map[string]interface{}{
			"displayName":     socket.Config.Name,
			"profileImageUrl": "",
			"customFields": map[string]interface{}{
				"workEmail": socket.Config.Email,
				"workPhone": socket.Config.Phone,
				"firstName": firstName,
				"lastName":  lastName,
			},
		},
	})

	url := fmt.Sprintf("https://api.%s/api/v2/webchat/guest/conversations", socket.Config.URI)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Println("err", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	defer resp.Body.Close()
	response, _ := ioutil.ReadAll(resp.Body)

	var conversation Response
	json.Unmarshal(response, &conversation)

	log.Println("conversation", conversation)

	socket.JWT = conversation.JWT
	socket.UserID = conversation.Member["id"]

	c, _, err := websocket.DefaultDialer.Dial(conversation.StreamURI, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	socket.SocketConn = c

	go func() {
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var msg Message
			json.Unmarshal(message, &msg)
			log.Println("recv: ", msg)

			if msg.TopicName != "channel.metadata" {
				socket.ConversationID = msg.Event.Conversation["id"]
				if msg.Event.Member["id"] == socket.UserID && msg.Event.Member["state"] == "CONNECTED" {
					socket.AgentID = msg.Event.Member["id"]
					socket.SendHistory()
				}
				if msg.Metadata["type"] == "message" || msg.Metadata["type"] == "typing-indicator" {
					if socket.UserID != msg.Event.Sender["id"] {
						if msg.Event.Body != "" {
							if !socket.Started {
								resp, _ := json.Marshal(map[string]string{
									"timestamp": msg.Event.Timestamp,
								})
								socket.StartHandler(messageType, resp)
								socket.Started = true
								log.Println("socket.Started", socket.Started)
							}
							resp, _ := json.Marshal(map[string]string{
								"type":      "text",
								"text":      msg.Event.Body,
								"timestamp": msg.Event.Timestamp,
							})
							socket.MessageHandler(messageType, resp)
						} else {
							resp, _ := json.Marshal(map[string]string{
								"type": "typing",
							})
							socket.MessageHandler(messageType, resp)
						}
					}
				}
				if msg.Metadata["type"] == "member-change" && msg.Event.Member["id"] == socket.UserID && msg.Event.Member["state"] == "DISCONNECTED" {
					resp, _ := json.Marshal(map[string]string{})
					socket.AgentLeftHandler(messageType, resp)
					socket.EndHandler(messageType, resp)
					socket.End()
					return
				}
			}
		}
	}()
}

// End ...
func (socket *SocketService) End() {
	log.Println("purecloud Ended ....")
	err := socket.SocketConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
	}
}

// Send ...
func (socket *SocketService) Send(msg string) {
	log.Println("Send message to purecloud", msg)

	body, _ := json.Marshal(map[string]interface{}{
		"body":     msg,
		"bodyType": "standard",
	})

	url := fmt.Sprintf("https://api.%s/api/v2/webchat/guest/conversations/%s/members/%s/messages", socket.Config.URI, socket.ConversationID, socket.UserID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer "+socket.JWT)
	client := &http.Client{}
	client.Do(req)
}

// SendHistory ...
func (socket *SocketService) SendHistory() {
	history := "--- bot history ---\n"
	for _, h := range socket.History {
		history += h + "\n"
	}
	history += "--- bot history ---\n"
	socket.Send(history)
}
