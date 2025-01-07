package edge

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/bincooo/emit.io"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func Authorize(session *emit.Session, ctx context.Context, scopeId, idToken, cookie string) (accessToken string, err error) {
	token, _ := jwt.Parse(idToken, func(token *jwt.Token) (zero interface{}, err error) { return })
	if token == nil {
		err = fmt.Errorf("invalid jwt")
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	challenge, verifier := genVerifier()

	rid := uuid.NewString()
	XAM := "Oid:00000000-0000-0000-591d-" + randString(12) + "@" + uuid.NewString()
	response, err := emit.ClientBuilder(session).
		Context(ctx).
		GET("https://login.microsoftonline.com/common/oauth2/v2.0/authorize").
		Query("client_id", claims["aud"].(string)).
		Query("scope", url.QueryEscape(fmt.Sprintf("%s/ChatAI.ReadWrite openid profile offline_access", scopeId))).
		Query("redirect_uri", url.QueryEscape("https://copilot.microsoft.com")).
		Query("client-request-id", rid).
		Query("response_mode", "fragment").
		Query("response_type", "code").
		Query("x-client-SKU", "msal.js.browser").
		Query("x-client-VER", "3.26.1").
		Query("client_info", "1").
		Query("code_challenge", challenge).
		Query("code_challenge_method", "S256").
		Query("prompt", "none").
		Query("login_hint", url.QueryEscape(claims["login_hint"].(string))).
		Query("X-AnchorMailbox", url.QueryEscape(XAM)).
		Query("nonce", uuid.NewString()).
		Query("state", base64.StdEncoding.EncodeToString([]byte("{\"id\":\""+uuid.NewString()+"\",\"meta\":{\"interactionType\":\"silent\"}}"))).
		Query("lw", "1").
		Query("coa", "1").
		Query("nopa", "2").
		Query("fl", "easi2_wld").
		Query("cobrandid", uuid.NewString()).
		Header("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7").
		Header("accept-encoding", "gzip, deflate, br, zstd").
		Header("accept-language", "en-US,en;q=0.9").
		Header("referer", "https://copilot.microsoft.com/").
		Header("user-agent", userAgent).
		Header("cookie", cookie).
		DoS(http.StatusFound)
	if err != nil {
		return
	}
	response.Body.Close()
	cookies := emit.GetCookies(response)

	location := response.Header.Get("Location")
	inst, err := url.Parse(location)
	if err != nil {
		return
	}

	query := inst.Query()
	response, err = emit.ClientBuilder(session).
		Context(ctx).
		GET("https://login.live.com/oauth20_authorize.srf").
		Query("client_id", claims["aud"].(string)).
		Query("scope", url.QueryEscape(fmt.Sprintf("%s/ChatAI.ReadWrite openid profile offline_access", scopeId))).
		Query("redirect_uri", url.QueryEscape("https://copilot.microsoft.com")).
		Query("response_type", "code").
		Query("state", query.Get("state")).
		Query("response_mode", "fragment").
		Query("nonce", query.Get("nonce")).
		Query("prompt", "none").
		Query("login_hint", url.QueryEscape(claims["login_hint"].(string))).
		Query("code_challenge", query.Get("code_challenge")).
		Query("code_challenge_method", "S256").
		Query("x-client-SKU", "msal.js.browser").
		Query("x-client-VER", "3.26.1").
		Query("uaid", query.Get("uaid")).
		Query("msproxy", "1").
		Query("issuer", "mso").
		Query("tenant", "common").
		Query("ui_locales", "en-US").
		Query("client_info", "1").
		Query("epct", query.Get("epct")).
		Query("jshs", "0").
		Query("nopa", "2").
		Query("fl", "easi2_wld").
		Query("cobrandid", query.Get("cobrandid")).
		Query("lw", "1").
		Query("coa", "1").
		Header("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7").
		Header("accept-encoding", "gzip, deflate, br, zstd").
		Header("accept-language", "en-US,en;q=0.9").
		Header("referer", "https://copilot.microsoft.com/").
		Header("user-agent", userAgent).
		Header("cookie", emit.MergeCookies(cookie, cookies)).
		DoS(http.StatusFound)
	if err != nil {
		return
	}
	response.Body.Close()
	location = strings.Replace(response.Header.Get("Location"), "copilot.microsoft.com/#", "copilot.microsoft.com/?", 1)
	inst, err = url.Parse(location)
	if err != nil {
		return
	}

	query = inst.Query()
	if message := query.Get("error_description"); message != "" {
		err = fmt.Errorf(message)
		return
	}

	accessToken, err = refreshToken0(session, ctx, rid, claims["aud"].(string), scopeId, query.Get("code"), verifier, XAM)
	return
}

func refreshToken0(session *emit.Session, ctx context.Context, rid, clientId, scopeId, code, verifier, XAM string) (accessToken string, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("client_id", clientId)
	_ = writer.WriteField("redirect_uri", "https://copilot.microsoft.com")
	_ = writer.WriteField("scope", fmt.Sprintf("%s/ChatAI.ReadWrite openid profile offline_access", scopeId))
	_ = writer.WriteField("code", code)
	_ = writer.WriteField("x-client-SKU", "msal.js.browser")
	_ = writer.WriteField("x-client-VER", "3.26.1")
	_ = writer.WriteField("x-ms-lib-capability", "retry-after, h429")
	_ = writer.WriteField("x-client-current-telemetry", "5|863,0,,,|,")
	_ = writer.WriteField("x-client-last-telemetry", "5|0|61,"+rid+"|bad_token|1,0")
	_ = writer.WriteField("code_verifier", verifier)
	_ = writer.WriteField("grant_type", "authorization_code")
	_ = writer.WriteField("client_info", "1")
	_ = writer.WriteField("client-request-id", rid)
	_ = writer.WriteField("X-AnchorMailbox", XAM)
	err = writer.Close()
	if err != nil {
		return
	}

	response, err := emit.ClientBuilder(session).
		Context(ctx).
		POST("https://login.microsoftonline.com/common/oauth2/v2.0/token").
		Header("accept-language", "en-US,en;q=0.9").
		Header("content-type", writer.FormDataContentType()).
		Header("origin", "https://copilot.microsoft.com").
		Header("referer", "https://copilot.microsoft.com/").
		Header("x-edge-shopping-flag", "1").
		Header("user-agent", userAgent).
		Buffer(body).
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		return
	}

	obj, err := emit.ToMap(response)
	if err != nil {
		return
	}

	if data, ok := obj["token_type"]; !ok || data != "Bearer" {
		err = fmt.Errorf("refresh failed")
		return
	}

	if data, ok := obj["scope"]; !ok || !strings.HasSuffix(data.(string), "/ChatAI.ReadWrite") {
		err = fmt.Errorf("refresh failed")
		return
	}

	accessToken = obj["refresh_token"].(string) + "|" + obj["access_token"].(string)
	return
}

func RefreshToken(session *emit.Session, ctx context.Context, token string) (accessToken string, err error) {
	split := strings.Split(token, "|")
	if len(split) < 3 {
		err = fmt.Errorf("refresh token is unauthorized")
		return
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("client_id", split[0])
	_ = writer.WriteField("redirect_uri", "https://copilot.microsoft.com")
	_ = writer.WriteField("scope", split[1]+"/ChatAI.ReadWrite openid profile offline_access")
	_ = writer.WriteField("grant_type", "refresh_token")
	_ = writer.WriteField("client_info", "1")
	_ = writer.WriteField("x-client-SKU", "msal.js.browser")
	_ = writer.WriteField("x-client-VER", "3.26.1")
	_ = writer.WriteField("x-ms-lib-capability", "retry-after, h429")
	_ = writer.WriteField("x-client-current-telemetry", "5|61,0,,,|,")
	_ = writer.WriteField("x-client-last-telemetry", "5|40|||0,0")
	_ = writer.WriteField("client-request-id", uuid.NewString())
	_ = writer.WriteField("refresh_token", strings.Join(split[2:], "|"))
	_ = writer.WriteField("X-AnchorMailbox", "Oid:00000000-0000-0000-591d-"+randString(12)+"@"+uuid.NewString())
	err = writer.Close()
	if err != nil {
		return
	}

	response, err := emit.ClientBuilder(session).
		Context(ctx).
		POST("https://login.microsoftonline.com/common/oauth2/v2.0/token").
		Header("accept-language", "en-US,en;q=0.9").
		Header("content-type", writer.FormDataContentType()).
		Header("origin", "https://copilot.microsoft.com").
		Header("referer", "https://copilot.microsoft.com/").
		Header("user-agent", userAgent).
		Header("x-edge-shopping-flag", "1").
		Buffer(body).
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		return
	}

	obj, err := emit.ToMap(response)
	if err != nil {
		return
	}

	if data, ok := obj["token_type"]; !ok || data != "Bearer" {
		err = fmt.Errorf("refresh failed")
		return
	}

	if data, ok := obj["scope"]; !ok || !strings.HasSuffix(data.(string), "/ChatAI.ReadWrite") {
		err = fmt.Errorf("refresh failed")
		return
	}

	accessToken = obj["refresh_token"].(string) + "|" + obj["access_token"].(string)
	return
}

func genVerifier() (challenge, verifier string) {
	data := make([]byte, 32)
	rand.Read(data)
	verifier = base64.StdEncoding.EncodeToString(data)
	verifier = strings.ReplaceAll(verifier, "=", "")
	verifier = strings.ReplaceAll(verifier, "+", "-")
	verifier = strings.ReplaceAll(verifier, "/", "_")

	data = []byte(verifier)
	sum256 := sha256.Sum256(data)
	challenge = base64.StdEncoding.EncodeToString(sum256[:])
	challenge = strings.ReplaceAll(challenge, "=", "")
	challenge = strings.ReplaceAll(challenge, "+", "-")
	challenge = strings.ReplaceAll(challenge, "/", "_")
	return
}
