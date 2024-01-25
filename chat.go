package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/RomiChan/websocket"
	"github.com/google/uuid"
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

func (ws *wsConn) Close() {
	ws.IsClose = true
	_ = ws.Conn.Close()
}

func NewDefaultOptions(cookie, middle string) (*Options, error) {
	var bu string
	var ws string
	if middle != "" {
		u, err := url.Parse(middle)
		if err != nil {
			return nil, &ChatError{"options", err}
		}

		bu = middle + "/turing/conversation/create"
		ws = "wss://" + u.Hostname() + "/sydney/ChatHub"
	} else {
		bu = DefaultCreate
		ws = DefaultChatHub
	}

	co := cookie
	if cookie != "" && !strings.Contains(cookie, "_U=") {
		co = "_U=" + cookie
	}

	return &Options{
		retry:  2,
		wss:    ws,
		create: bu,
		middle: middle,
		model:  ModelCreative,
		headers: map[string]string{
			"Cookie": co,
		},
	}, nil
}

// 设置本地代理地址
func (opts *Options) Proxies(proxies string) *Options {
	opts.proxies = proxies
	return opts
}

// 设置对话模式
func (opts *Options) Model(model string) *Options {
	opts.model = model
	return opts
}

// 温度调节 0.0~1.0, Sydney 模式生效
func (opts *Options) Temperature(temperature float32) *Options {
	opts.temperature = temperature
	return opts
}

// topic警告是否作为错误返回，默认为false
func (opts *Options) TopicToE(flag bool) *Options {
	opts.topicToE = flag
	return opts
}

// 文档模式
func (opts *Options) Notebook(flag bool) *Options {
	opts.notebook = flag
	return opts
}

// 插件
func (opts *Options) Plugins(plugins ...string) *Options {
	opts.plugins = plugins
	return opts
}

func (opts *Options) KievAuth(kievRPSSecAuth, rwBf string) *Options {
	opts.kievRPSSecAuth = kievRPSSecAuth
	opts.rwBf = rwBf
	return opts
}

// 创建会话实例
func New(opts *Options) Chat {
	if opts == nil {
		return Chat{
			Options: Options{
				middle: "https://copilot.microsoft.com",
			},
		}
	}
	has := func(key string) bool {
		for k, _ := range opts.headers {
			if strings.ToLower(k) == key {
				return true
			}
		}
		return false
	}

	for k, v := range H {
		if !has(strings.ToLower(k)) {
			opts.headers[k] = v
		}
	}
	if opts.middle == "" {
		opts.middle = "https://copilot.microsoft.com"
	}
	return Chat{Options: *opts}
}

func (c *Chat) GetSession() Conversation {
	if c.session == nil {
		return Conversation{}
	} else {
		return *c.session
	}
}

// 对话并回复
//
// ctx Context 控制器，promp string 当前对话，image KBlob 图片信息，previousMessages []map[string]string 历史记录
//
// previousMessages:
//
//	[
//		{
//			"author": "user",
//			"text": "hi"
//		},
//		{
//			"author": "bot",
//			"text": "Hello, this is Bing. I am a chat mode ..."
//		}
//	]
func (c *Chat) Reply(ctx context.Context, prompt string, image *KBlob, previousMessages []ChatMessage) (chan ChatResponse, error) {
	c.mu.Lock()
	if c.session == nil || c.session.ConversationId == "" || c.model == ModelSydney {
		count := 1
	label:
		conv, err := c.newConversation()
		if err != nil {
			if count < c.retry {
				count++
				goto label
			}
			c.mu.Unlock()
			return nil, &ChatError{"conversation", err}
		}
		c.session = conv
	}

	h, err := c.newHub(c.model, *c.session, prompt, image, previousMessages)
	if err != nil {
		c.mu.Unlock()
		return nil, &ChatError{"data", err}
	}

	hub := map[string]any{
		"arguments":    []any{h},
		"invocationId": strconv.Itoa(c.session.invocationId),
		"target":       "chat",
		"type":         4,
	}

	conn, err := c.newConn()
	if err != nil {
		c.mu.Unlock()
		return nil, &ChatError{"conn", err}
	}

	marshal, err := json.Marshal(hub)
	if err != nil {
		c.mu.Unlock()
		return nil, &ChatError{"marshal", err}
	}

	err = conn.WriteMessage(websocket.TextMessage, append(marshal, delimiter))
	if err != nil {
		c.mu.Unlock()
		return nil, &ChatError{"writeMessage", err}
	}

	message := make(chan ChatResponse)
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
func (c *Chat) resolve(ctx context.Context, conn *wsConn, message chan ChatResponse) {
	defer close(message)
	defer c.mu.Unlock()
	defer conn.Close()

	normal := false

	h := func() bool {
		// 轮询回复消息
		_, marshal, err := conn.ReadMessage()
		if err != nil {
			conn.IsClose = true
			message <- ChatResponse{
				Error: &ChatError{
					Action:  "resolve",
					Message: err,
				},
			}
			return true
		}

		// 是心跳应答
		if bytes.Equal(marshal, ping) {
			return false
		}

		var response partialResponse
		slice := bytes.Split(marshal, []byte{delimiter})
		err = json.Unmarshal(slice[0], &response)
		if err != nil {
			message <- ChatResponse{
				Error: &ChatError{"resolve", err},
			}
			return false
		}

		result := ChatResponse{
			RawData: slice[0],
		}

		// 结束本次应答
		if response.Type == 2 {
			if response.Item.Result.Value != "Success" {
				result.Error = &ChatError{"resolve", errors.New(response.Item.Result.Message)}
				message <- result
			}
			conn.IsClose = true
			if t := response.Item.T; t != nil {
				result.T = &struct {
					Max  int
					Used int
				}{t.Max, t.Used}
			}

			if messages := response.Item.Messages; !normal && messages != nil && len(*messages) > 0 {
				topicMessage := findTopicMessage(*messages)
				if topicMessage != "" {
					if c.topicToE {
						result.Error = &ChatError{"resolve", errors.New(topicMessage)}
					} else {
						c.session.invocationId++
						result.Text = "\n" + topicMessage
					}
				}
			} else {
				c.session.invocationId++
			}
			message <- result
			return true
		}

		// 消息响应失败
		if response.Type == 3 {
			conn.IsClose = true
			result.Error = &ChatError{"resolve", errors.New(response.Error)}
			message <- result
			return true
		}

		if len(response.Args) == 0 {
			return false
		}

		// 处理消息
		args0 := response.Args[0]
		if args0.Messages == nil || len(*args0.Messages) == 0 {
			return false
		}

		m := (*args0.Messages)[0]
		if m.MessageType != "" && strings.Contains("InternalSearchQuery,InternalSearchResult,InternalLoaderMessage", m.MessageType) {
			return false
		}

		if containsTopicToE(m.Text) {
			if !normal && c.topicToE {
				result.Error = &ChatError{"resolve", errors.New(m.Text)}
				message <- result
				return true
			}

			if normal {
				return true
			}
		}

		result.Text = m.Text
		// 有正常输出，则忽略TopicMessage警告
		if len(result.Text) > 0 {
			normal = true
		}
		message <- result
		return false
	}

	for {
		select {
		case <-ctx.Done():
			message <- ChatResponse{
				Error: &ChatError{"timeout", errors.New("resolve timeout")},
			}
			return
		default:
			if h() {
				return
			}
		}
	}
}

func containsTopicToE(value string) bool {
	blocks := []string{
		"That’s on me",
		"different topic",
	}
	for _, block := range blocks {
		if strings.Contains(value, block) {
			return true
		}
	}
	return false
}

func findTopicMessage(messages []struct {
	Author     string `json:"author"`
	Text       string `json:"text"`
	Type       string `json:"messageType"`
	SpokenText string `json:"spokenText"`
}) string {
	messageL := len(messages)
	if messageL == 0 {
		return ""
	}
	for i := messageL - 1; i >= 0; i-- {
		if msg := messages[i]; msg.Author == "bot" && msg.Type != "Disengaged" {
			if msg.SpokenText != "" {
				return msg.SpokenText
			}
		}
	}
	return ""
}

// 创建websocket
func (c *Chat) newConn() (*wsConn, error) {
	header := c.newHeader()
	header.Add("accept-language", "en-US,en;q=0.9")
	// header.Add("origin", "https://edgeservices.bing.com")
	header.Add("origin", "https://copilot.microsoft.com")
	header.Add("host", "sydney.bing.com")

	dialer := websocket.DefaultDialer
	if c.proxies != "" {
		purl, err := url.Parse(c.proxies)
		if err != nil {
			return nil, err
		}
		dialer = &websocket.Dialer{
			Proxy:            http.ProxyURL(purl),
			HandshakeTimeout: 45 * time.Second,
		}
	}

	if c.session == nil {
		return nil, errors.New("the conversation value was unexpectedly nil")
	}
	conn, _, err := dialer.Dial(c.wss+"?sec_access_token="+url.QueryEscape(c.session.accessToken), header)
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
func (c *Chat) newHub(model string, conv Conversation, prompt string, image *KBlob, previousMessages []ChatMessage) (map[string]any, error) {
	var hub map[string]any
	if c.notebook {
		if err := json.Unmarshal(nbkHub, &hub); err != nil {
			return nil, err
		}
	} else if err := json.Unmarshal(chatHub, &hub); err != nil {
		return nil, err
	}

	messageId := uuid.NewString()
	if model == "" {
		model = ModelCreative
	}
	if model == ModelSydney {
		var tone string
		if c.temperature > .6 {
			tone = ModelCreative
		} else if c.temperature > .3 {
			tone = ModelBalanced
		} else {
			tone = ModelPrecise
		}
		messageTypes := hub["allowedMessageTypes"].([]any)
		h := func(str string) func(any) bool {
			return func(item any) bool {
				return item == str
			}
		}
		messageTypes = del(messageTypes, h("SearchQuery"))
		messageTypes = del(messageTypes, h("RenderCardRequest"))
		messageTypes = del(messageTypes, h("InternalSearchQuery"))
		messageTypes = del(messageTypes, h("InternalSearchResult"))
		hub["allowedMessageTypes"] = messageTypes
		hub["tone"] = tone

		if !c.notebook {
			hub["sliceIds"] = sliceIds
		}
	} else {
		hub["tone"] = model
	}

	plugins, err := c.LoadPlugins(c.plugins...)
	if err != nil {
		return nil, err
	}

	if len(plugins) > 0 {
		_plugins := make([]map[string]any, 0)
		for _, plugin := range plugins {
			_plugins = append(_plugins, map[string]any{
				"id":       plugin,
				"category": 1,
			})
		}
		hub["plugins"] = _plugins
	}

	hub["traceId"] = conv.traceId
	hub["requestId"] = messageId
	hub["conversationId"] = conv.ConversationId
	hub["participant"] = kv{
		"id": conv.ClientId,
	}

	message, ok := hub["message"].(map[string]any)
	if !ok {
		return nil, errors.New("failed to reflect 'message'")
	}

	message["timestamp"] = time.Now().Format("2006-01-02T15:04:05+08:00")
	message["requestId"] = messageId
	message["messageId"] = messageId
	if image != nil {
		message["imageUrl"] = "https://copilot.microsoft.com/images/blob?bcid=" + image.ProcessedBlobId
		message["originalImageUrl"] = "https://copilot.microsoft.com/images/blob?bcid=" + image.BlobId
	}
	message["text"] = prompt

	if conv.invocationId == 0 || model == ModelSydney {
		// 处理历史消息
		hub["isStartOfSession"] = true
		if len(previousMessages) > 0 {
			hub["previousMessages"] = previousMessages
			conv.invocationId = len(previousMessages) / 2
		} else {
			delete(hub, "previousMessages")
		}
	} else if model != "Sydney" {
		hub["isStartOfSession"] = false
		delete(hub, "previousMessages")
	}
	return hub, nil
}

// 删除会话
func (c *Chat) Delete() error {
	conversationId := c.session.ConversationId
	if conversationId == "" {
		return nil
	}

	request, err := http.NewRequest(http.MethodGet, c.create+"?conversationId="+c.session.ConversationId, nil)
	if err != nil {
		return &ChatError{"delete", err}
	}

	request.Header = c.newHeader()
	client, err := c.newClient()
	if err != nil {
		return &ChatError{"delete", err}
	}

	r, err := client.Do(request)
	if err != nil {
		return &ChatError{"delete", err}
	}

	if r.StatusCode != http.StatusOK {
		return &ChatError{"delete", errors.New(r.Status)}
	}

	marshal, err := io.ReadAll(r.Body)
	if err != nil {
		return &ChatError{"delete", err}
	}

	var conv Conversation
	if err = json.Unmarshal(marshal, &conv); err != nil {
		return &ChatError{"delete", err}
	}

	authorization := r.Header.Get("X-Sydney-Conversationsignature")
	params := map[string]any{
		"conversationId": c.session.ConversationId,
		"optionsSets": []string{
			"autosave",
			"savemem",
			"uprofupd",
			"uprofgen",
		},
		"source": "cib",
	}
	requestUrl := c.middle
	marshal, _ = json.Marshal(params)
	if c.middle == "" || c.middle == "https://www.bing.com" {
		requestUrl = "https://sydney.bing.com"
	}
	request, err = http.NewRequest(http.MethodPost, requestUrl+"/sydney/DeleteSingleConversation", bytes.NewBuffer(marshal))
	if err != nil {
		return &ChatError{"delete", err}
	}
	request.Header = c.newHeader()
	request.Header.Set("Authorization", "Bearer "+authorization)
	request.Header.Set("Content-Type", "application/json")
	client, err = c.newClient()
	if err != nil {
		return &ChatError{"delete", err}
	}
	r, err = client.Do(request)
	if err != nil {
		return &ChatError{"delete", err}
	}
	_, _ = io.ReadAll(r.Body)
	return nil
}

// 获取bing插件ID。需要包含Search，否则无效。
// 可用插件 Shop 、Instacart、OpenTable、Klarna、Search、Kayak
func (c *Chat) LoadPlugins(names ...string) (plugins []string, err error) {
	if len(names) == 0 {
		return make([]string, 0), nil
	}

	middle := c.middle
	if strings.Contains(middle, "https://copilot.microsoft.com") {
		middle = "https://www.bing.com"
	}

	request, err := http.NewRequest(http.MethodGet, middle+"/codex/plugins/available/get", nil)
	if err != nil {
		return nil, err
	}

	request.Header = c.newHeader()
	client, err := c.newClient()
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

	var result map[string]any
	if e := json.Unmarshal(marshal, &result); e != nil {
		return nil, e
	}

	if result["IsSuccess"] == true {
		if value, ok := result["Value"].([]interface{}); ok {
			for _, item := range value {
				object := item.(map[string]interface{})
				if contains(names, object["Name"].(string)) {
					plugins = append(plugins, object["Id"].(string))
				}
			}
		}
	}
	return
}

// 创建会话
func (c *Chat) newConversation() (*Conversation, error) {
	request, err := http.NewRequest(http.MethodGet, c.create+"?bundleVersion="+Version, nil)
	if err != nil {
		return nil, err
	}

	request.Header = c.newHeader()
	client, err := c.newClient()
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

	conv.traceId = strings.ReplaceAll(uuid.NewString(), "-", "")
	conv.invocationId = 0
	conv.accessToken = r.Header.Get("X-Sydney-Encryptedconversationsignature")
	cookies := r.Header.Values("Set-Cookie")
	for _, cookie := range cookies {
		if cookie[:5] == "MUID=" {
			if muId := strings.Split(cookie, "; ")[0][5:]; muId != "" {
				c.muId = muId
			}
			break
		}
	}

	return &conv, nil
}

func (c *Chat) newClient() (*http.Client, error) {
	client := http.DefaultClient
	if c.proxies != "" {
		purl, err := url.Parse(c.proxies)
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

func (c *Chat) newHeader() http.Header {
	var h = make(http.Header)
	for k, v := range c.headers {
		if strings.ToLower(k) == "cookie" {
			if v == "_U=" {
				v = ""
			}
			if c.kievRPSSecAuth != "" && !strings.Contains(v, "KievRPSSecAuth=") {
				v += "; KievRPSSecAuth=" + c.kievRPSSecAuth
			}
			if c.rwBf != "" && !strings.Contains(v, "_RwBf=") {
				v += "; _RwBf=" + c.rwBf
			}
			if c.muId != "" {
				v += "; MUID=" + c.muId
			}
		}
		if v != "" {
			h.Set(k, v)
		}
	}
	return h
}

func del[T any](slice []T, condition func(item T) bool) []T {
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

// 判断切片是否包含子元素
func contains[T comparable](slice []T, t T) bool {
	return containFor(slice, func(item T) bool {
		return item == t
	})
}

// 判断切片是否包含子元素， condition：自定义判断规则
func containFor[T comparable](slice []T, condition func(item T) bool) bool {
	if len(slice) == 0 {
		return false
	}

	for idx := 0; idx < len(slice); idx++ {
		if condition(slice[idx]) {
			return true
		}
	}
	return false
}
