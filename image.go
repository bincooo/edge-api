package edge

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "image/jpeg"
	_ "image/png"
)

const (
	maxPixels float64 = 360000.0
	ua                = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"
)

func (c *Chat) LoadImage(file string) (*KBlob, error) {
	var (
		dataBytes []byte
		err       error
	)

	// base64
	if strings.HasPrefix(file, "data:image/") {
		pos := strings.Index(file, ";")
		if pos == -1 {
			return nil, &ChatError{"image", errors.New("invalid base64 url")}
		}

		file = file[pos+1:]
		if !strings.HasPrefix(file, "base64,") {
			return nil, &ChatError{"image", errors.New("invalid base64 url")}
		}

		dataBytes, err = base64.StdEncoding.DecodeString(file[7:])
		if err != nil {
			return nil, &ChatError{"image", err}
		}

	} else if strings.HasPrefix(file, "http") {
		// url
		response, e := http.Get(file)
		if e != nil {
			return nil, &ChatError{"image", e}
		}
		defer response.Body.Close()

		dataBytes, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, &ChatError{"image", err}
		}

	} else {
		// local file
		dataBytes, err = os.ReadFile(file)
		if os.IsNotExist(err) {
			return nil, &ChatError{"image", err}
		}
	}

	i, _, err := image.Decode(bytes.NewReader(dataBytes))
	if err != nil {
		return nil, &ChatError{"image", err}
	}

	b := i.Bounds()
	width := float64(b.Max.X)
	height := float64(b.Max.Y)

	pixels := maxPixels / (width * height)
	if pixels < 1 {
		rate := math.Sqrt(pixels)
		width *= rate
		height *= rate
	}

	buf := new(bytes.Buffer)
	result := resize.Resize(uint(width), uint(height), i, resize.Lanczos3)
	if err = jpeg.Encode(buf, result, nil); err != nil {
		return nil, &ChatError{"image", err}
	}

	dataBytes = buf.Bytes()
	base64Image := base64.StdEncoding.EncodeToString(dataBytes)
	return c.uploadBase64(base64Image)
}

// 上传文件。 middle: 服务器地址， proxies: 本地代理， base64Image: 图片base64编码
func (c *Chat) uploadBase64(base64Image string) (kb *KBlob, err error) {
	body := new(bytes.Buffer)

	if c.session == nil {
		c.session, err = c.newConversation()
		if err != nil {
			return nil, &ChatError{"conversation", err}
		}
	}

	w := multipart.NewWriter(body)
	_ = w.WriteField("knowledgeRequest", `{
			"imageInfo": {},
			"knowledgeRequest": {
			"invokedSkills": [
			  	"ImageById"
			],
			"subscriptionId": "Bing.Chat.Multimodal",
			"invokedSkillsRequestData": {
			  	"enableFaceBlur": true
			},
			"convoData": {
			  	"convoid": "`+c.session.ConversationId+`",
			  	"convotone": "Balanced"
			}
	  	}
	}`)
	_ = w.WriteField("imageBase64", base64Image)
	_ = w.Close()

	request, err := http.NewRequest(http.MethodPost, c.middle+"/images/kblob", body)
	if err != nil {
		return kb, &ChatError{"image", err}
	}

	headers := c.newHeader()
	headers.Set("Content-Type", w.FormDataContentType())
	if strings.Contains(c.middle, "www.bing.com") {
		headers.Set("origin", "https://www.bing.com")
		headers.Set("referer", "https://www.bing.com/search?q=Bing+AI")
	}
	if strings.Contains(c.middle, "copilot.microsoft.com") {
		headers.Set("origin", "https://copilot.microsoft.com")
		headers.Set("referer", "https://copilot.microsoft.com")
	}

	request.Header = headers
	client := http.DefaultClient
	if c.proxies != "" {
		var curl *url.URL
		curl, err = url.Parse(c.proxies)
		if err != nil {
			return kb, &ChatError{"proxies", err}
		}
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(curl),
			},
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return kb, &ChatError{"image", err}
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return kb, &ChatError{"image", errors.New(response.Status)}
	}

	marshal, err := io.ReadAll(response.Body)
	if err != nil {
		return kb, &ChatError{"image", err}
	}

	if err = json.Unmarshal(marshal, &kb); err != nil {
		return kb, &ChatError{"image", err}
	}
	if kb.ProcessedBlobId == "" {
		kb.ProcessedBlobId = kb.BlobId
	}
	return kb, nil
}
