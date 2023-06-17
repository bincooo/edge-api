package edge

import (
	"sync"
)

type Options struct {
	Headers   map[string]string
	Retry     int
	WebSock   string
	CreateURL string
	Model     string
}

type Chat struct {
	Options
	mu sync.Mutex

	Session Conversation
}

type Conversation struct {
	ConversationId string `json:"conversationId"`
	ClientId       string `json:"clientId"`
	Signature      string `json:"conversationSignature"`

	Result struct {
		Value   string      `json:"value"`
		Message interface{} `json:"message"`
	} `json:"result"`

	TraceId      string `json:"-"`
	InvocationId int    `json:"-"`
}

type PartialResponse struct {
	Error error `json:"-"`

	Type         int    `json:"type"`
	InvocationId string `json:"invocationId"`

	Arguments []struct {
		RequestId string `json:"requestId"`

		Throttling *struct {
			Max  int `json:"maxNumUserMessagesInConversation"`
			Used int `json:"numUserMessagesInConversation"`
		} `json:"throttling"`

		Messages *[]struct {
			Text        string `json:"text"`
			MessageType string `json:"messageType"`
		} `json:"messages"`
	} `json:"arguments"`

	Item *struct {
		Result struct {
			Message string `json:"message"`
			Value   string `json:"value"`
		} `json:"result"`

		Messages *[]struct {
			Author string `json:"author"`
			Text   string `json:"text"`
		} `json:"messages"`
	} `json:"item"`

	Text string `json:"-"`
}
