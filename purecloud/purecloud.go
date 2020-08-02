package purecloud

import (
	"encoding/json"
)

// Slug ...
const Slug = "PureCloud"

// PureCloud ...
type PureCloud struct {
	socketService *SocketService
}

// New ...
func New(config json.RawMessage, history []string, messageHandler func(int, json.RawMessage), startHandler func(int, json.RawMessage),
	agentLeftHandler func(int, json.RawMessage), endHandler func(int, json.RawMessage), failedHandler func(int, json.RawMessage)) *PureCloud {

	conf := Config{}
	json.Unmarshal(config, &conf)

	socketService := &SocketService{
		Config:           conf,
		History:          history,
		MessageHandler:   messageHandler,
		StartHandler:     startHandler,
		AgentLeftHandler: agentLeftHandler,
		EndHandler:       endHandler,
		FailedHandler:    failedHandler,
	}
	return &PureCloud{
		socketService: socketService,
	}
}

// Start ...
func (pc *PureCloud) Start() {
	pc.socketService.Start()
}

// End ...
func (pc *PureCloud) End() {
	pc.socketService.End()
}

// Send ...
func (pc *PureCloud) Send(message json.RawMessage) {
	var data UserMessage
	json.Unmarshal(message, &data)
	if data.Event == "message" {
		pc.socketService.Send(data.Data.Text)
	}
}
