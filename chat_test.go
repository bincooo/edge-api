package edge

import (
	"context"
	"testing"

	"github.com/bincooo/emit.io"

	_ "embed"
)

var (
	//go:embed message.txt
	query string

	accessToken = "xxx"
)

func TestConversation(t *testing.T) {
	session, err := emit.NewSession("http://127.0.0.1:7890", nil)
	if err != nil {
		t.Fatal(err)
	}

	conversationId, err := CreateConversation(session, context.TODO(), accessToken)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(conversationId)
}

func TestChat(t *testing.T) {
	session, err := emit.NewSession("http://127.0.0.1:7890", nil)
	if err != nil {
		t.Fatal(err)
	}

	conversationId, err := CreateConversation(session, context.TODO(), accessToken)
	if err != nil {
		t.Fatal(err)
	}

	message, err := Chat(session, context.TODO(), accessToken, conversationId, query)
	if err != nil {
		t.Fatal(err)
	}

	for {
		chunk, ok := <-message
		if !ok {
			break
		}

		if chunk[0] == 1 {
			t.Fatalf("%s", chunk[1:])
			return
		}

		t.Logf("%s", chunk[1:])
	}

	err = DeleteConversation(session, context.TODO(), conversationId, accessToken)
	if err != nil {
		t.Fatal(err)
	}
}
