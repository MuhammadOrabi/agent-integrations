package manager

import (
	"agent-integrations/purecloud"
	"encoding/json"
)

// Client ...
type Client struct {
	Driver           string
	Config           json.RawMessage
	History          []string
	MessageHandler   func(int, json.RawMessage)
	StartHandler     func(int, json.RawMessage)
	AgentLeftHandler func(int, json.RawMessage)
	EndHandler       func(int, json.RawMessage)
	FailedHandler    func(int, json.RawMessage)
}

// Integeration ...
type Integeration interface {
	Start()
	End()
	Send(json.RawMessage)
}

// NewClient ...
func NewClient(c Client) Integeration {
	if c.Driver == purecloud.Slug {
		return purecloud.New(c.Config, c.History, c.MessageHandler, c.StartHandler, c.AgentLeftHandler, c.EndHandler, c.FailedHandler)
	}
	return nil
}
