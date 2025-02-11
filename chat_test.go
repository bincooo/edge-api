package edge

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/bincooo/emit.io"

	_ "embed"
)

var (
	//go:embed message.txt
	query string

	clientId     = "14638111-3389-403d-b206-a6a71d9f8f16"
	scopeId_1    = "140e65af-45d1-4427-bf08-3e7295db6836"
	refreshToken = "M.C550_BAY.0.U.-Cp6bPDPPLqjby32k9ppLvdIUL0abRG5i2VnTTaZLtuh3rLECAcA!FycqhJWCkp31Xuv8sv!pODHjTty7G7vq7tk6VNxcWWL9kh3Ovo586YCUYecmbY!NmqYiD44!hkseqZKvFhOVDKy34zvXnBv8LVhvqdJ!A6bA7lifYJR5SMw7q4PDE5VISeeinpowJyR1GBZBmQyMLVki6nOtFCnv*X4hZhHyaGkhxhDDQHBAON4PtX5x7KoPvX389l1YfTFyifSNqzsP8JYZxIPsjfcXOmyYsGz7l9fU!BlYsbCh*i5MMk4vgAhc97h1VsSTz8sHGuCiXG9rHVxsFcGnghSOGhEA*VMUxCoUQwne5ino0QoN"
)

func TestConversation(t *testing.T) {
	session, err := emit.NewSession("http://127.0.0.1:7890", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	accessToken, err := RefreshToken(session, context.TODO(), clientId, scopeId_1, refreshToken)
	if err != nil {
		t.Fatal(err)
	}

	conversationId, err := CreateConversation(session, context.TODO(), strings.Split(accessToken, "|")[1])
	if err != nil {
		t.Fatal(err)
	}
	t.Log(conversationId)

	buffer, err := os.ReadFile("/Users/bincooo/Desktop/Screenshot 2024-12-31 at 17.20.01.png")
	if err != nil {
		t.Fatal(err)
	}

	uri, err := Attachments(session, context.TODO(), buffer, strings.Split(accessToken, "|")[1])
	if err != nil {
		t.Fatal(err)
	}
	t.Log(uri)
}

func TestChat(t *testing.T) {
	session, err := emit.NewSession("http://127.0.0.1:7890", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	accessToken, err := RefreshToken(session, context.TODO(), clientId, scopeId_1, refreshToken)
	if err != nil {
		t.Fatal(err)
	}

	conversationId, err := CreateConversation(session, context.TODO(), strings.Split(accessToken, "|")[1])
	if err != nil {
		t.Fatal(err)
	}

	buffer, err := os.ReadFile("/Users/bincooo/Desktop/Screenshot 2024-12-31 at 17.20.01.png")
	if err != nil {
		t.Fatal(err)
	}

	uri, err := Attachments(session, context.TODO(), buffer, strings.Split(accessToken, "|")[1])
	if err != nil {
		t.Fatal(err)
	}

	message, err := Chat(session, context.TODO(), strings.Split(accessToken, "|")[1], conversationId, "", "", "图里写了什么？", uri, 0)
	if err != nil {
		t.Fatal(err)
	}

	text := ""
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
		var obj map[string]interface{}
		err = json.Unmarshal(chunk[1:], &obj)
		if err != nil {
			t.Fatal(err)
		}
		if obj["event"] == "appendText" {
			text += obj["text"].(string)
		}
	}

	t.Log(text)
	err = DeleteConversation(session, context.TODO(), conversationId, strings.Split(accessToken, "|")[1])
	if err != nil {
		t.Fatal(err)
	}
}
