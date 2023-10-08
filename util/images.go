package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var H = map[string]string{
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.55",
}

type KBlob struct {
	BlobId          string `json:"blobId"`
	ProcessedBlobId string `json:"processedBlobId"`
}

func UploadImage(baseUrl, proxy, image string) (KBlob, error) {
	fBytes, err := os.ReadFile(image)
	if err != nil {
		return KBlob{}, err
	}
	return UploadImageBase64(baseUrl, proxy, base64.StdEncoding.EncodeToString(fBytes))
}

// 上传文件。 baseUrl: 服务器地址， proxy: 本地代理， base64Image: 图片base64编码
func UploadImageBase64(baseUrl, proxy, base64Image string) (KBlob, error) {
	var kb KBlob
	body := new(bytes.Buffer)

	w := multipart.NewWriter(body)
	_ = w.WriteField("knowledgeRequest", `{"imageInfo":{},"knowledgeRequest":{"invokedSkills":["ImageById"],"subscriptionId":"Bing.Chat.Multimodal","invokedSkillsRequestData":{"enableFaceBlur":true},"convoData":{"convoid":"","convotone":"Creative"}}}`)
	_ = w.WriteField("imageBase64", base64Image)
	_ = w.Close()

	request, err := http.NewRequest(http.MethodPost, baseUrl+"/images/kblob", body)
	if err != nil {
		return kb, err
	}

	request.Header.Set("Content-Type", w.FormDataContentType())
	if strings.Contains(baseUrl, "www.bing.com") {
		request.Header.Set("origin", "https://www.bing.com")
		request.Header.Set("referer", "https://www.bing.com/search?q=Bing+AI")
	}
	client := http.DefaultClient
	if proxy != "" {
		curl, e := url.Parse(proxy)
		if e != nil {
			return kb, e
		}
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(curl),
			},
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return kb, err
	}

	if response.StatusCode != http.StatusOK {
		return kb, errors.New(response.Status)
	}

	marshal, err := io.ReadAll(response.Body)
	if err != nil {
		return kb, err
	}

	if err = json.Unmarshal(marshal, &kb); err != nil {
		return kb, err
	}
	return kb, nil
}

// 从对话中提取图片路径
func ParseImage(prompt string) (image, result string) {
	regexCompile := regexp.MustCompile(`\{image:[^}]+}\n`)
	imageSchema := regexCompile.FindString(prompt)
	if imageSchema == "" {
		return "", prompt
	}
	result = strings.Replace(prompt, imageSchema, "", -1)
	imageSchema = strings.TrimPrefix(imageSchema, "{image:")
	imageSchema = strings.TrimSuffix(imageSchema, "}\n")
	return strings.TrimSpace(imageSchema), result
}

// 解析上传图片信息: {blob:blobId#processedBlobId}\n
func ParseKBlob(prompt string) (*KBlob, string) {
	if prompt != "" {
		return nil, prompt
	}

	regexCompile := regexp.MustCompile(`\{blob:[^}]+}\n`)
	blobSchema := regexCompile.FindString(prompt)
	if blobSchema == "" {
		return nil, prompt
	}

	result := strings.Replace(prompt, blobSchema, "", -1)
	blobSchema = strings.TrimPrefix(blobSchema, "{blob:")
	blobSchema = strings.TrimSuffix(blobSchema, "}\n")
	slice := strings.Split(blobSchema, "#")
	return &KBlob{slice[0], slice[1]}, result
}
