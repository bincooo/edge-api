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

	// client_id | scope_id | refresh_token
	refreshToken = "14638111-3389-403d-b206-a6a71d9f8f16|140e65af-45d1-4427-bf08-3e7295db6836|M.C550_BAY.0.U.-ChfPypJYai2JPBV0wFJz075iLODHa3P9HLGZeFbEkYfWznXt0V5l0YDaXgBekptKYSuvOAcO*1wURFBpNpqK!kTyxU4jdENtPLuUaNEGKrDGPgU1ZJI9aQk7zs7yCcvEjRCldfMSH9CSzBXxeN6jc2kCz1gAI2rR92!S0DSvlZfJjQRupsXg0Zd3*O386hkne4or6sJkkeVz7VBTX13J7lb0S9SWU*j563PhVfv4Njt686Ghh*WSzvYlFkAQfuQBDPv16AjT9d*ISJtQC8jl*JE8GYWVuKeV!tIhFr89CfDWLpNkU3VzU4bVGfAh!JI8OYkoJ!XhcQWb88S3emtkJwk7VGYn5mS07PzDuR!IHqVh"
)

func TestConversation(t *testing.T) {
	session, err := emit.NewSession("http://127.0.0.1:7890", nil)
	if err != nil {
		t.Fatal(err)
	}

	accessToken, err := RefreshToken(session, context.TODO(), refreshToken)
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

	accessToken, err := RefreshToken(session, context.TODO(), refreshToken)
	if err != nil {
		t.Fatal(err)
	}

	conversationId, err := CreateConversation(session, context.TODO(), accessToken)
	if err != nil {
		t.Fatal(err)
	}

	message, err := Chat(session, context.TODO(), accessToken, conversationId, "", query, "")
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
