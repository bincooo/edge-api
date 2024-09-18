package edge

import (
	"fmt"
	"github.com/bincooo/emit.io"
	"github.com/google/uuid"
	"math/rand"
	"sync"
	"time"
)

type Options struct {
	cookies        string  //
	temperature    float32 // 温度调节：通过不同温度调节对话模式
	kievRPSSecAuth string  //
	rwBf           string  //
	topicToE       bool    // topic警告是否作为错误返回
	notebook       bool    // 文档模式
	compose        bool    // 混合模式 ？ 效用待测
	composeObj     ComposeObj
	optionSets     []interface{} // 自定义参数
	plugins        []string      // 插件

	model   string // 对话模式
	retry   int    // 重试次数
	proxies string // 本地代理
	middle  string // 中间转发地址
	muId    string // 设备Id？
	wss     string // 对话链接
	create  string // 创建会话链接
}

type ComposeObj struct {
	Fmt    string
	Length string
	Tone   string
}

type Chat struct {
	Options
	mu sync.Mutex

	session *Conversation
	blob    *KBlob
	client  *emit.Session
}

func (c *Chat) KBlob(blob *KBlob) {
	c.blob = blob
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
	Error string `json:"error"`

	Type         int    `json:"type"`
	InvocationId string `json:"invocationId"`

	Args []struct {
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

		T *struct {
			Max  int `json:"maxNumUserMessagesInConversation"`
			Used int `json:"numUserMessagesInConversation"`
		} `json:"throttling"`
	} `json:"item"`
}

type ChatMessage = map[string]interface{}

type ChatResponse struct {
	Text    string
	Error   *ChatError
	RawData []byte

	T *struct {
		Max  int
		Used int
	}
}

type ChatError struct {
	Action  string
	Message error
}

func (kb KBlob) String() string {
	return fmt.Sprintf(`{ BlobId: %s, ProcessedBlobId: %s}`, kb.BlobId, kb.ProcessedBlobId)
}

func (c ChatError) Error() string {
	return fmt.Sprintf("[edge-api::%s] %v", c.Action, c.Message)
}

func BuildMessage(messageType string, text ...string) ChatMessage {
	message, locale := extractArgs(text)
	switch messageType {
	case "Internal", "CurrentWebpageContextRequest":
	default:
		messageType = ""
	}

	result := ChatMessage{
		"text":   message,
		"author": "user",
	}

	if messageType != "" {
		result["messageType"] = messageType
	}

	result["locale"] = locale
	return result
}

func BuildSwitchMessage(role string, text ...string) ChatMessage {
	switch role {
	case "bot":
		return BuildBotMessage(text...)
	default:
		return BuildUserMessage(text...)
	}
}

func BuildUserMessage(text ...string) ChatMessage {
	message, locale := extractArgs(text)
	return ChatMessage{
		"text":         message,
		"author":       "user",
		"messageId":    "local-gen-" + uuid.NewString(),
		"requestId":    uuid.NewString(),
		"responseType": 6,
		"isFromCache":  true,
		"genStream":    true,
		"locale":       locale,
	}
}

func BuildBotMessage(text ...string) ChatMessage {
	message, locale := extractArgs(text)
	return ChatMessage{
		"text":          message,
		"author":        "bot",
		"invocation":    "hint(Copilot_language=\"中文\")",
		"messageId":     uuid.NewString(),
		"offense":       "None",
		"contentOrigin": "CachedEntry",
		"responseType":  6,
		"isFromCache":   true,
		"genStream":     true,
		"locale":        locale,
	}
}

func BuildPageMessage(text ...string) ChatMessage {
	message, locale := extractArgs(text)
	return ChatMessage{
		"author":      "user",
		"description": message,
		"contextType": "WebPage",
		"messageType": "Context",
		"sourceName":  "histories.txt",
		"sourceUrl":   "file:///Users/" + randStr(5) + "/histories.txt",
		"locale":      locale,
	}
}

func extractArgs(text []string) (string, string) {
	locale := "en-US"
	textL := len(text)
	if textL == 0 {
		return "", locale
	}

	if textL > 1 {
		locale = text[1]
	}

	return text[0], locale
}

func randStr(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	bytes := make([]rune, n)
	for i := range bytes {
		bytes[i] = runes[r.Intn(len(runes))]
	}
	return string(bytes)
}
