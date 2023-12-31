package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/RomiChan/websocket"
	"github.com/bincooo/edge-api/util"
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
	"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1 Edg/120.0.0.0",
}

type kv = map[string]string

type wsConn struct {
	*websocket.Conn
	IsClose bool
}

func NewDefaultOptions(cookie, agency string) (Options, error) {
	var bu string
	var ws string
	if agency != "" {
		u, err := url.Parse(agency)
		if err != nil {
			return Options{}, err
		}

		bu = agency + "/turing/conversation/create"
		ws = "wss://" + u.Hostname() + "/sydney/ChatHub"
	} else {
		bu = DefaultCreate
		ws = DefaultChatHub
	}

	co := cookie
	if cookie != "" && !strings.Contains(cookie, "_U=") {
		co = "_U=" + cookie
	}

	return Options{
		Retry:     2,
		WebSock:   ws,
		CreateURL: bu,
		Model:     Creative,
		Headers: map[string]string{
			"Cookie": co,
		},
	}, nil
}

func New(opts Options) *Chat {
	has := func(key string) bool {
		for k, _ := range opts.Headers {
			if strings.ToLower(k) == key {
				return true
			}
		}
		return false
	}

	for k, v := range H {
		if !has(strings.ToLower(k)) {
			opts.Headers[k] = v
		}
	}
	if opts.agency == "" {
		opts.agency = "https://www.bing.com"
	}
	chat := Chat{Options: opts}
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

	h, err := c.newHub(c.Model, c.Session, prompt, previousMessages)
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
	marshal = bytes.ReplaceAll(marshal, []byte("\n"), []byte(""))
	marshal = bytes.ReplaceAll(marshal, []byte(" "), []byte(""))
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
		const s5 = 5 * time.Second
		t := time.Now().Add(s5)
		for {
			if conn.IsClose {
				return
			}
			// 5秒执行一次心跳
			if time.Now().After(t) {
				t = time.Now().Add(s5)
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

		response.RawData = slice[0]
		// 结束本次应答
		if response.Type == 2 {
			if response.Item.Result.Value != "Success" {
				response.Error = errors.New("消息响应失败：" + response.Item.Result.Message)
				message <- response
			}
			_ = conn.Close()
			conn.IsClose = true
			c.Session.InvocationId++
			message <- response
			return true
		}

		if response.Type == 3 {
			response.Error = errors.New("消息响应失败：" + response.InnerError)
			message <- response
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

		if m.HiddenText == "" {
			response.Text = m.Text
		}
		message <- response
		return false
	}

	for {
		select {
		case <-ctx.Done():
			message <- PartialResponse{
				Error: errors.New("resolve timeout"),
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
	header := c.initHeader()
	header.Add("accept-language", "en-US,en;q=0.9")
	// header.Add("origin", "https://edgeservices.bing.com")
	header.Add("origin", "https://copilot.microsoft.com")
	header.Add("host", "sydney.bing.com")

	dialer := websocket.DefaultDialer
	if c.Proxy != "" {
		purl, e := url.Parse(c.Proxy)
		if e != nil {
			return nil, e
		}
		dialer = &websocket.Dialer{
			Proxy:            http.ProxyURL(purl),
			HandshakeTimeout: 45 * time.Second,
		}
	}

	ustr := c.WebSock
	if c.Session.AccessToken != "" {
		ustr += "?sec_access_token=" + url.QueryEscape(c.Session.AccessToken)
	}
	conn, _, err := dialer.Dial(ustr, header)
	if err != nil {
		return nil, err
	}

	if e := conn.WriteMessage(websocket.TextMessage, schema); e != nil {
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
func (c *Chat) newHub(model string, conv Conversation, prompt string, previousMessages []map[string]string) (map[string]any, error) {
	var hub map[string]any
	if err := json.Unmarshal(chatHub, &hub); err != nil {
		return nil, err
	}

	messageId := uuid.NewString()
	if model == "" {
		model = Creative
	}
	if model == Sydney {
		var tone string
		if c.Temperature > .6 {
			tone = Creative
		} else if c.Temperature > .3 {
			tone = Balanced
		} else {
			tone = Precise
		}
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
		hub["tone"] = tone
	} else {
		hub["sliceIds"] = sliceIds
		hub["tone"] = model
	}

	hub["traceId"] = conv.TraceId
	// hub["conversationSignature"] = conv.Signature

	hub["requestId"] = messageId
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
	if blob, tex := util.ParseKBlob(prompt); blob != nil {
		message["imageUrl"] = "https://www.bing.com/images/blob?bcid=" + blob.ProcessedBlobId
		message["originalImageUrl"] = "https://www.bing.com/images/blob?bcid=" + blob.BlobId
		prompt = tex
	}
	message["text"] = prompt

	if conv.InvocationId == 0 || model == Sydney {
		// 处理历史消息
		hub["isStartOfSession"] = true
		if len(previousMessages) > 0 {
			for _, previousMessage := range previousMessages {
				if previousMessage["author"] != "user" {
					continue
				}
				text := previousMessage["text"]
				if blob, tex := util.ParseKBlob(text); blob != nil {
					previousMessage["imageUrl"] = "https://www.bing.com/images/blob?bcid=" + blob.ProcessedBlobId
					previousMessage["originalImageUrl"] = "https://www.bing.com/images/blob?bcid=" + blob.BlobId
					previousMessage["text"] = tex
				}
			}
			hub["previousMessages"] = previousMessages
			conv.InvocationId = len(previousMessages) / 2
		} else {
			delete(hub, "previousMessages")
		}
	} else if model != "Sydney" {
		delete(hub, "previousMessages")
	}
	return hub, nil
}

// 删除会话
func (c *Chat) Delete() error {
	conversationId := c.Session.ConversationId
	if conversationId == "" {
		return nil
	}

	request, err := http.NewRequest(http.MethodGet, c.CreateURL+"?conversationId="+c.Session.ConversationId, nil)
	if err != nil {
		return err
	}

	request.Header = c.initHeader()
	client, err := c.initClient()
	if err != nil {
		return err
	}

	r, err := client.Do(request)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errors.New(r.Status)
	}

	marshal, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var conv Conversation
	if err = json.Unmarshal(marshal, &conv); err != nil {
		return err
	}

	authorization := r.Header.Get("X-Sydney-Conversationsignature")
	params := map[string]any{
		"conversationId": c.Session.ConversationId,
		"optionsSets": []string{
			"autosave",
			"savemem",
			"uprofupd",
			"uprofgen",
		},
		"source": "cib",
	}
	requestUrl := c.agency
	marshal, _ = json.Marshal(params)
	if c.agency == "" || c.agency == "https://www.bing.com" {
		requestUrl = "https://sydney.bing.com"
	}
	request, err = http.NewRequest(http.MethodPost, requestUrl+"/sydney/DeleteSingleConversation", bytes.NewBuffer(marshal))
	if err != nil {
		return err
	}
	request.Header = c.initHeader()
	request.Header.Set("Authorization", "Bearer "+authorization)
	request.Header.Set("Content-Type", "application/json")
	client, err = c.initClient()
	if err != nil {
		return err
	}
	r, err = client.Do(request)
	if err != nil {
		return err
	}
	readBytes, _ := io.ReadAll(r.Body)
	logrus.Info("delete session: ", string(readBytes))
	return nil
}

// 创建会话
func (c *Chat) newConversation() (*Conversation, error) {
	request, err := http.NewRequest(http.MethodGet, c.CreateURL+"?bundleVersion="+Version, nil)
	if err != nil {
		return nil, err
	}

	request.Header = c.initHeader()
	client, err := c.initClient()
	if err != nil {
		return nil, err
	}

	r, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
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
	conv.AccessToken = r.Header.Get("X-Sydney-Encryptedconversationsignature")
	cookies := r.Header.Values("Set-Cookie")
	for _, cookie := range cookies {
		if cookie[:5] == "MUID=" {
			if muid := strings.Split(cookie, "; ")[0][5:]; muid != "" {
				c.MUID = muid
			}
			break
		}
	}

	return &conv, nil
}

func (c *Chat) initClient() (*http.Client, error) {
	client := http.DefaultClient
	if c.Proxy != "" {
		purl, err := url.Parse(c.Proxy)
		if err != nil {
			return nil, err
		}
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(purl),
			},
		}
	}
	return client, nil
}

func (c *Chat) initHeader() http.Header {
	var h = make(http.Header)
	for k, v := range c.Headers {
		if strings.ToLower(k) == "cookie" {
			if v == "_U=" {
				v = ""
			}
			if c.KievRPSSecAuth != "" && !strings.Contains(v, "KievRPSSecAuth=") {
				v += "; KievRPSSecAuth=" + c.KievRPSSecAuth
			}
			if c.RwBf != "" && !strings.Contains(v, "_RwBf=") {
				v += "; _RwBf=" + c.RwBf
			}
			if c.MUID != "" {
				v += "; MUID=" + c.MUID
			}
		}
		if v != "" {
			h.Set(k, v)
		}
	}
	return h
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
