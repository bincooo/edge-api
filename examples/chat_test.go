package main

import (
	_ "embed"

	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"strings"
	"testing"
)

const (
	cookie = "xxx"

	KievAuth = "xxx"
	RwBf     = "xxx"
)

//go:embed histories.txt
var histories string

// 前置引导
var pMessages = []edge.ChatMessage{
	edge.BuildUserMessage("你好"),
	edge.BuildBotMessage("你好，这是必应。我可以用中文和你聊天，也可以帮你做一些有趣的事情，比如写诗，编程，创作歌曲，角色扮演等等。你想让我做什么呢？😊"),
	edge.BuildUserMessage("你能做什么"),
	edge.BuildBotMessage("我能做很多有趣和有用的事情，比如：\n\n- 和你聊天，了解你的兴趣和爱好，根据你的要求扮演一些有趣的角色或故事。\n- 从当前网页中的内容回答问题。\n- 描述你上传的图片，告诉你图片里有什么，或者画一幅你想要的图画。\n\n你想让我试试哪一项呢？😊"),
}

func TestPlugins(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}

	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Notebook(true))

	r, err := chat.LoadPlugins(context.Background(), edge.PluginSearch, edge.PluginOpenTable)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestSearchMessages(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	options.KievAuth(KievAuth, RwBf).Notebook(true)
	// Sydney 模式需要自行维护历史对话
	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelSydney).
		Temperature(.9).
		TopicToE(true))
	chat.Plugins(edge.PluginSearch)
	response, err := chat.Reply(context.Background(), "你对gemini1.5 flush版本怎么看", pMessages)
	if err != nil {
		t.Fatal(err)
	}
	resolve(t, response)
}

func TestClassification(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	//options.KievAuth(KievAuth, RwBf)

	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Notebook(true))
	partialResponse, err := chat.Reply(context.Background(), strings.Replace(template1, "{{prompt}}", "解析网页内容", -1), nil)
	if err != nil {
		t.Fatal(err)
	}

	response := resolve(t, partialResponse)
	t.Log("response: ", response)
	if strings.Contains(response, "wqre") {
		t.Log("assert: website-crawler____getWebsiteContent")

	} else if strings.Contains(response, "sdfa") {
		t.Log("assert: website-crawler____getWeather")
	} else {
		t.Log("assert: other")
		return
	}

	partialResponse, err = chat.Reply(context.Background(), strings.Replace(template2, "{{description}}", "查看对话记录，提取有效的Schema属性", -1), nil)
	if err != nil {
		t.Fatal(err)
	}

	response = resolve(t, partialResponse)

	t.Log("response: ", response)
}

func TestNsfwMessages(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	options.KievAuth(KievAuth, RwBf).Notebook(true)
	// Sydney 模式需要自行维护历史对话
	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelSydney).
		Temperature(.9).
		TopicToE(true))
	//chat.Compose(true, edge.ComposeObj{
	//	Fmt:    "paragraph",
	//	Length: "long",
	//	Tone:   "Fond, Easy, Obedient",
	//})
	t.Logf("cookie登陆状态: %v", chat.IsLogin(context.Background()))
	pMessages = append(pMessages, edge.BuildPageMessage(histories))

	//text := "继续"
	text := "今天心情如何？"
	// text := "给主人看看"
	// text := "居然是蕾丝边的，可以送给我么"
	//text := "拜托你了"
	fmt.Println("You: ", text)
	partialResponse, err := chat.Reply(context.Background(), text, pMessages)
	if err != nil {
		t.Fatal(err)
	}
	response := resolve(t, partialResponse)
	t.Log(response)
}

func TestLinkMessages(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	options.KievAuth(KievAuth, RwBf)
	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Temperature(1.0).
		TopicToE(true))

	text := "为我提供一些golang的文章"
	fmt.Println("You: ", text)
	partialResponse, err := chat.Reply(context.Background(), text, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := resolve(t, partialResponse)
	t.Log(response)
}

func Test_image(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	options.KievAuth(KievAuth, RwBf)
	options.Proxies("socks5://127.0.0.1:7890")
	options.Model(edge.ModelSydney)
	chat := edge.New(options)
	file := "https://www.1micro.top/alist/d/blob.jpg"
	kb, err := chat.LoadImage(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(kb)
	chat.KBlob(kb)
	partialResponse, err := chat.Reply(context.Background(), "请你使用json代码块中文描述这张图片，不必说明直接输出结果", nil)
	if err != nil {
		t.Fatal(err)
	}

	text := resolve(t, partialResponse)
	t.Log(text)
}

func resolve(t *testing.T, partialResponse chan edge.ChatResponse) string {
	msg := ""
	for {
		message, ok := <-partialResponse
		if !ok {
			return msg
		}

		if message.Error != nil {
			t.Fatal(message.Error)
		}

		if len(message.Text) > 0 {
			msg = message.Text
		}

		fmt.Println(message.Text)
		fmt.Println("===============")
		if message.T != nil {
			t.Logf("used: %d / %d\n", message.T.Max, message.T.Used)
		}
	}
}
