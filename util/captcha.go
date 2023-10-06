package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
)

var (
	baseURL = "https://666102.201666.xyz"
)

func init() {
	_ = godotenv.Load()
	baseURL = LoadEnvVar("BING_CAPTCHA_URL", baseURL)
}

func LoadEnvVar(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

// https://github.com/ikechan8370/chatgpt-plugin/blob/64f2699b2210cf7f4b655e66d09cfc0b133361cf/utils/bingCaptcha.js#L64
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
