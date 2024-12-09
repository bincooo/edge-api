package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/RomiChan/websocket"
	"github.com/bincooo/emit.io"
)

var (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1.1 Safari/605.1.15"

	// 懒得定义了
	H = []byte("{\"event\":\"setOptions\",\"ads\":{\"supportedTypes\":[\"text\",\"propertyPromotion\",\"tourActivity\",\"product\",\"multimedia\"]}}")
	D = []byte("{\"event\":\"done\"")
	E = []byte("{\"event\":\"error\"")
)

type Msg struct {
	Event          string `json:"event"`
	Mode           string `json:"mode"`
	ConversationId string `json:"conversationId"`
	Content        []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func CreateConversation(session *emit.Session, ctx context.Context) (conversationId string, err error) {
	response, err := emit.ClientBuilder(session).
		Context(ctx).
		POST("https://copilot.microsoft.com/c/api/conversations").
		Header("accept-language", "en-US,en;q=0.9").
		Header("origin", "https://copilot.microsoft.com").
		Header("user-agent", userAgent).
		Header("referer", "https://copilot.microsoft.com/chats").
		Header("x-search-uilang", "en-us").
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		return
	}

	defer response.Body.Close()
	obj, err := emit.ToMap(response)
	if err != nil {
		return
	}

	conversationId = obj["id"].(string)
	return
}

func Chat(session *emit.Session, ctx context.Context, conversationId, query string) (message chan []byte, err error) {
	conn, _, err := emit.SocketBuilder(session).
		Context(ctx).
		URL("wss://copilot.microsoft.com/c/api/chat").
		Query("api-version", "2").
		Header("origin", "https://copilot.microsoft.com").
		Header("user-agent", userAgent).
		DoS(http.StatusSwitchingProtocols)
	if err != nil {
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, H)
	if err != nil {
		return
	}

	request := Msg{
		Event:          "send",
		Mode:           "chat",
		ConversationId: conversationId,
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{{Type: "text", Text: query}},
	}
	hexBytes, err := json.Marshal(request)
	if err != nil {
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, hexBytes)
	if err != nil {
		return
	}

	message = make(chan []byte)
	go resolve(conn, message)
	return
}

func resolve(conn *websocket.Conn, message chan []byte) {
	defer close(message)
	defer conn.Close()

	for {
		_, chunk, err := conn.ReadMessage()
		if err != nil {
			message <- messageBuffer(1, err)
			return
		}

		if bytes.HasPrefix(chunk, D) {
			return
		}

		if bytes.HasPrefix(chunk, E) {
			message <- messageBuffer(1, chunk)
		}

		message <- messageBuffer(0, chunk)
	}
}

func messageBuffer(magic byte, o interface{}) (buffer []byte) {
	buffer = []byte{magic}
	if o == nil {
		return
	}

	if err, ok := o.(error); ok {
		buffer = append([]byte{magic}, []byte(err.Error())...)
		return
	}

	if hex, ok := o.(string); ok {
		buffer = append([]byte{magic}, []byte(hex)...)
		return
	}

	if hex, ok := o.([]byte); ok {
		buffer = append(buffer, hex...)
	}
	return
}
