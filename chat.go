package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/RomiChan/websocket"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var H = map[string]string{
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0",
}

type kv = map[string]string

type wsConn struct {
	*websocket.Conn
	IsClose bool
}

func New(cookie, agency string) (*Chat, error) {
	var bu string
	var ws string
	if agency != "" {
		u, err := url.Parse(agency)
		if err != nil {
			return nil, err
		}

		bu = agency + "/turing/conversation/create"
		ws = "wss://" + u.Hostname() + "/sydney/ChatHub"
	} else {
		bu = DefaultCreate
		ws = DefaultChatHub
	}

	co := cookie
	if !strings.Contains(cookie, "_U=") {
		co = "_U=" + cookie
	}
	return NewChat(Options{
		Retry:     2,
		WebSock:   ws,
		CreateURL: bu,
		Model:     Creative,
		Headers: map[string]string{
			"Cookie": co,
		},
	}), nil
}

func NewChat(opt Options) *Chat {
	has := func(key string) bool {
		for k, _ := range opt.Headers {
			if strings.ToLower(k) == key {
				return true
			}
		}
		return false
	}

	for k, v := range H {
		if !has(strings.ToLower(k)) {
			opt.Headers[k] = v
		}
	}

	chat := Chat{Options: opt}
	return &chat
}

func (c *Chat) Reply(ctx context.Context, prompt string, previousMessages []map[string]string) (chan PartialResponse, error) {
	c.mu.Lock()
	if c.Session.ConversationId == "" || c.Model == Sydney {
		cnt := 1
	label:
		conv, err := c.newConversation()
		if err != nil {
			if cnt < c.Retry {
				cnt++
				goto label
			}
			c.mu.Unlock()
			return nil, err
		}
		c.Session = *conv
	}

	h, err := newHub(c.Model, c.Session, prompt, previousMessages)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	hub := map[string]any{
		"arguments":    []any{h},
		"invocationId": strconv.Itoa(c.Session.InvocationId),
		"target":       "chat",
		"type":         4,
	}

	conn, err := c.newConn()
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	marshal, err := json.Marshal(hub)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	err = conn.WriteMessage(websocket.TextMessage, bytes.Join([][]byte{marshal, []byte(Delimiter)}, []byte{}))
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	message := make(chan PartialResponse)
	go c.resolve(ctx, conn, message)
	go func() {
		const s15 = 15 * time.Second
		t := time.Now().Add(s15)
		for {
			if conn.IsClose {
				return
			}
			// 15秒执行一次心跳
			if time.Now().After(t) {
				t = time.Now().Add(s15)
				err = conn.WriteMessage(websocket.TextMessage, ping)
				if err != nil {
					return
				}
			}
			time.Sleep(time.Second)
		}
	}()
	return message, nil
}

// 解析回复信息
func (c *Chat) resolve(ctx context.Context, conn *wsConn, message chan PartialResponse) {
	defer close(message)
	defer c.mu.Unlock()
	handle := func() bool {
		// 轮询回复消息
		_, marshal, err := conn.ReadMessage()
		if err != nil {
			conn.IsClose = true
			message <- PartialResponse{
				Error: err,
			}
			return true
		}

		// 是心跳应答
		if bytes.Equal(marshal, ping) {
			return false
		}

		//fmt.Println(string(marshal))
		var response PartialResponse
		slice := bytes.Split(marshal, []byte(Delimiter))

		logrus.Debug(string(slice[0]))
		err = json.Unmarshal(slice[0], &response)
		if err != nil {
			message <- PartialResponse{
				Error: err,
			}
			return false
		}

		// 结束本次应答
		if response.Type == 2 {
			if response.Item.Result.Value != "Success" {
				response.Error = errors.New("消息响应失败：" + response.Item.Result.Message)
				message <- response
			}
			_ = conn.Close()
			conn.IsClose = true
			c.Session.InvocationId++
			//messages := response.Item.Messages
			//if messages != nil && len(*messages) > 1 {
			//	var texts []string
			//	for _, item := range *messages {
			//		if item.Author == "bot" {
			//			texts = append(texts, item.Text)
			//			if item.Text == "" && item.SpokenText != "" {
			//				texts = append(texts, item.SpokenText)
			//			}
			//		}
			//	}
			//	response.Text = strings.Join(texts, "\n")
			message <- response
			//}
			return true
		}

		if len(response.Arguments) == 0 {
			return false
		}

		// 处理消息
		argument := response.Arguments[0]
		if argument.Messages == nil || len(*argument.Messages) == 0 {
			return false
		}

		m := (*argument.Messages)[0]
		if m.MessageType == "InternalSearchQuery" ||
			m.MessageType == "InternalSearchResult" ||
			m.MessageType == "InternalLoaderMessage" {
			return false
		}

		response.RowData = marshal
		response.Text = m.Text
		message <- response
		return false
	}

	for {
		select {
		case <-ctx.Done():
			message <- PartialResponse{
				Error: errors.New("请求超时"),
			}
			_ = conn.Close()
			return
		default:
			if handle() {
				return
			}
		}
	}
}

// 创建websocket
func (c *Chat) newConn() (*wsConn, error) {
	header := http.Header{}
	for k, v := range c.Headers {
		header.Add(k, v)
	}

	conn, _, err := websocket.DefaultDialer.Dial(c.WebSock, header)
	if err != nil {
		return nil, err
	}

	if e := conn.WriteMessage(websocket.TextMessage, Schema); e != nil {
		return nil, e
	}

	if _, _, e := conn.ReadMessage(); e != nil {
		return nil, e
	}

	if e := conn.WriteMessage(websocket.TextMessage, ping); e != nil {
		return nil, e
	}

	return &wsConn{conn, false}, nil
}

// 构建对接参数
func newHub(model string, conv Conversation, prompt string, previousMessages []map[string]string) (map[string]any, error) {
	var hub map[string]any
	if err := json.Unmarshal(chatHub, &hub); err != nil {
		return nil, err
	}

	messageId := uuid.NewString()
	if model == Sydney {
		// delete(hub, "allowedMessageTypes")
		amt := hub["allowedMessageTypes"].([]any)
		h := func(str string) func(any) bool {
			return func(item any) bool {
				return item == str
			}
		}
		amt = deleteItem(amt, h("SearchQuery"))
		amt = deleteItem(amt, h("RenderCardRequest"))
		amt = deleteItem(amt, h("InternalSearchQuery"))
		amt = deleteItem(amt, h("InternalSearchResult"))
		hub["allowedMessageTypes"] = amt
		hub["sliceIds"] = sliceIds //sSliceIds
		hub["tone"] = Creative
	} else {
		hub["sliceIds"] = sliceIds
		hub["tone"] = model
	}

	hub["traceId"] = conv.TraceId
	hub["requestId"] = messageId
	hub["conversationSignature"] = conv.Signature
	hub["conversationId"] = conv.ConversationId
	hub["participant"] = kv{
		"id": conv.ClientId,
	}

	message, ok := hub["message"].(map[string]any)
	if !ok {
		return nil, errors.New("reflect `message` fail")
	}

	message["timestamp"] = time.Now().Format("2006-01-02T15:04:05+08:00")
	message["requestId"] = messageId
	message["messageId"] = messageId
	message["text"] = prompt

	if conv.InvocationId == 0 || model == Sydney {
		hub["isStartOfSession"] = true
		hub["previousMessages"] = previousMessages
		if len(previousMessages) > 0 {
			conv.InvocationId = len(previousMessages) / 2
		}
	} else if model != "Sydney" {
		delete(hub, "previousMessages")
	}
	return hub, nil
}

// 创建会话
func (c *Chat) newConversation() (*Conversation, error) {
	request, err := http.NewRequest(http.MethodGet, c.CreateURL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range c.Headers {
		request.Header.Add(k, v)
	}

	r, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	marshal, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var conv Conversation
	if e := json.Unmarshal(marshal, &conv); e != nil {
		return nil, e
	}

	if c.TraceId != "" {
		conv.TraceId = c.TraceId
	} else {
		conv.TraceId = strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	conv.InvocationId = 0
	return &conv, nil
}

func deleteItem[T any](slice []T, condition func(item T) bool) []T {
	if len(slice) == 0 {
		return slice
	}

	for idx, element := range slice {
		if condition(element) {
			return append(slice[:idx], slice[idx+1:]...)
		}
	}
	return slice
}
