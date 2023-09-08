package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const (
	baseURL = "http://bingcaptcha.ikechan8370.com/bing"
)

func SolveCaptcha(token string) error {
	params := map[string]string{
		"_U": token,
	}
	marshal, err := json.Marshal(params)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(marshal))
	request.Header.Add("Content-Type", "application/json")
	r, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	marshal, err = io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var result map[string]any
	if e := json.Unmarshal(marshal, &result); e != nil {
		return e
	}

	if success, ok := result["success"]; ok {
		if success.(bool) {
			return nil
		} else {
			return errors.New("自动人机验证失败") //errors.New(result["statusText"].(string))
		}
	} else {
		return errors.New("自动人机验证失败")
	}
}
