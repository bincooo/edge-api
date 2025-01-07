package edge

import (
	"context"
	"github.com/bincooo/emit.io"
	"github.com/bogdanfinn/tls-client/profiles"
	"testing"
)

const (
	scopeId = "xxx"
	idToken = "xxx"
	cookie  = "xxx"
)

func TestAuthorize(t *testing.T) {
	session, err := emit.NewSession("http://127.0.0.1:7890", false, nil, emit.Ja3Helper(emit.Echo{RandomTLSExtension: true, HelloID: profiles.Chrome_124}, 10))
	if err != nil {
		t.Fatal(err)
	}

	accessToken, err := Authorize(session, context.TODO(), scopeId, idToken, cookie)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(accessToken)
}
