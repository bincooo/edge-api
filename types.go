package edge

import (
	"fmt"
	"sync"
)

type Options struct {
	headers        map[string]string // 默认请求头
	temperature    float32           // 温度调节：通过不同温度调节对话模式
	kievRPSSecAuth string            //
	rwBf           string            //
	topicToE       bool              // topic警告是否作为错误返回

	model   string // 对话模式
	retry   int    // 重试次数
	proxies string // 本地代理
	middle  string // 中间转发地址
	muId    string // 设备Id？
	wss     string // 对话链接
	create  string // 创建会话链接
}

type Chat struct {
	Options
	mu sync.Mutex

	session *Conversation
}

type KBlob struct {
	BlobId          string `json:"blobId"`
	ProcessedBlobId string `json:"processedBlobId"`
}

type Conversation struct {
	ConversationId string `json:"conversationId"`
	ClientId       string `json:"clientId"`

	Result struct {
		Value   string      `json:"value"`
		Message interface{} `json:"message"`
	} `json:"result"`

	traceId      string
	invocationId int
	accessToken  string
}

type partialResponse struct {
	InnerError string `json:"error"`

	Type         int    `json:"type"`
	InvocationId string `json:"invocationId"`

	Arguments []struct {
		RequestId string `json:"requestId"`
		Messages  *[]struct {
			Text        string `json:"text"`
			HiddenText  string `json:"hiddenText"`
			MessageType string `json:"messageType"`
		} `json:"messages"`
	} `json:"arguments"`

	Item *struct {
		Result struct {
			Message string `json:"message"`
			Value   string `json:"value"`
		} `json:"result"`

		Messages *[]struct {
			Author     string `json:"author"`
			Text       string `json:"text"`
			Type       string `json:"messageType"`
			SpokenText string `json:"spokenText"`
		} `json:"messages"`

		Throttling *struct {
			Max  int `json:"maxNumUserMessagesInConversation"`
			Used int `json:"numUserMessagesInConversation"`
		} `json:"throttling"`
	} `json:"item"`
}

type ChatMessage map[string]string

func (c ChatMessage) PushImage(image *KBlob) {
	if image == nil {
		return
	}
	c["imageUrl"] = "https://copilot.microsoft.com/images/blob?bcid=" + image.ProcessedBlobId
	c["originalImageUrl"] = "https://copilot.microsoft.com/images/blob?bcid=" + image.BlobId
}

type ChatResponse struct {
	Text    string
	Error   *ChatError
	RawData []byte
}

type ChatError struct {
	Action  string
	Message error
}

func (c ChatError) Error() string {
	return fmt.Sprintf("[EDGE-API::%s] %v", c.Action, c.Message)
}
