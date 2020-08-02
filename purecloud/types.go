package purecloud

// Config -> integeration config
type Config struct {
	OrganizationID string
	DeploymentID   string
	QueueName      string
	URI            string
	Name           string
	Phone          string
	Email          string
}

// Response -> init conversation response
type Response struct {
	StreamURI string            `json:"eventStreamUri"`
	JWT       string            `json:"jwt"`
	Member    map[string]string `json:"member"`
}

// EventBody -> message event body
type EventBody struct {
	ID           string            `json:"id"`
	Body         string            `json:"body"`
	BodyType     string            `json:"bodyType"`
	Timestamp    string            `json:"timestamp"`
	Conversation map[string]string `json:"conversation"`
	Sender       map[string]string `json:"sender"`
	Member       map[string]string `json:"member"`
}

// Message -> incoming message
type Message struct {
	TopicName string            `json:"topicName"`
	Version   string            `json:"version"`
	Event     EventBody         `json:"eventBody"`
	Metadata  map[string]string `json:"metadata"`
}

// UserMessageBody ...
type UserMessageBody struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
// UserMessage ...
type UserMessage struct {
	Event string          `json:"event"`
	Data  UserMessageBody `json:"data"`
}
