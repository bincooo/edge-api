package edge

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"github.com/bincooo/emit.io"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	_ "image/jpeg"
	_ "image/png"
)

const (
	maxPixels float64 = 360000.0
)

func (c *Chat) LoadImage(ctx context.Context, file string) (*KBlob, error) {
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
	return c.uploadBase64(ctx, base64Image)
}

// 上传文件。 middle: 服务器地址， proxies: 本地代理， base64Image: 图片base64编码
func (c *Chat) uploadBase64(ctx context.Context, base64Image string) (kb *KBlob, err error) {
	buffer := new(bytes.Buffer)

	if c.session == nil {
		c.session, err = c.newConversation(nil)
		if err != nil {
			return nil, &ChatError{"conversation", err}
		}
	}

	w := multipart.NewWriter(buffer)
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

	builder := emit.ClientBuilder(c.client).
		Context(ctx).
		Proxies(c.proxies).
		POST(c.middle+"/images/kblob").
		Header("Content-Type", w.FormDataContentType()).
		Header("cookie", c.extCookies()).
		Header("user-agent", userAgent)
	if strings.Contains(c.middle, "www.bing.com") {
		builder.Header("origin", "https://www.bing.com")
		builder.Header("referer", "https://www.bing.com/search?q=Bing+AI")
	}
	if strings.Contains(c.middle, "copilot.microsoft.com") {
		builder.Header("origin", "https://copilot.microsoft.com")
		builder.Header("referer", "https://copilot.microsoft.com")
	}

	response, err := builder.
		Buffer(buffer).
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		return kb, &ChatError{"image", err}
	}
	defer response.Body.Close()

	if err = emit.ToObject(response, &kb); err != nil {
		return kb, &ChatError{"image", err}
	}

	if kb.ProcessedBlobId == "" {
		kb.ProcessedBlobId = kb.BlobId
	}

	return kb, nil
}
