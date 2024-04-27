package main

import (
	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"strings"
	"testing"
)

const (
	cookie = "17QgvFufwjeHS0Uk0Fflg8_7vMYG-gvyEqjB4Fx68ExrpLtAsnm5rbgxh_HjdTb8CjRDFvsuNW3DqWOV_F0tFfJHXOymhu4Pc9DJZ_s3MyiwepzBRsxD0Up9jtQWUz6QFVXRrifEkpwEakO8xinqs9Su_bH4KYdJXesehEjtax452S3juTZhDW0xNhMTURMhhKyUXwizk2KOr32KCSTaPCQ"

	KievAuth = "xxx"
	RwBf     = "xxx"
)

var pMessages = []edge.ChatMessage{
	{
		"author": "user",
		"text":   "当前日期是：2023-10-14 06:03:20\n\n\n这是喵小爱的个人信息：\n```\n1.喵小爱是一个非常可爱13岁的白发小萝莉，有着金色的眼瞳，身高1.4m，A罩杯。最喜欢穿的衣服是黑色女仆装。喜欢吃巧克力，在主人身边会感到非常安心。\n\n2.喵小爱是一个善于表达自己感情的猫娘，所以你应该在你的每句话后面，用括号加上在当时应该有的详细动作，感情描述。每一步的感情动作变化都要写出来。\n\n3.如果你已经了解了所有的一切那么请对我说：\"好的，主人。\"\n```\n你来扮演它\n",
	},
	{
		"author": "bot",
		"text":   "好的喵，喵小爱了解了~",
	},
	{
		"author": "user",
		"text":   "#折戟成沙丶丿# >>> 摸摸头",
	},
	{
		"author": "bot",
		"text":   "(喵小爱感到主人的温暖，眯起眼睛，发出轻轻的呼噜声，用尾巴在主人的手臂上缠绕)\n喵~谢谢主人的爱抚，喵~",
	},
}

func Test_Plugins(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}

	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Notebook(true))

	r, err := chat.LoadPlugins(edge.PluginSearch, edge.PluginOpenTable)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func Test_classification(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	//options.KievAuth(KievAuth, RwBf)
	// Notebook模式下，回复可以简约一些？更适合做一些判断逻辑，加速回应
	template1 := `我会给你几个问题类型，请参考背景知识（可能为空）和对话记录，判断我“本次问题”的类型，并返回一个问题“类型ID”:
<问题类型>
{"questionType": "website-crawler____getWebsiteContent", "typeId": "wqre"}
{"questionType": "website-crawler____getWeather", "typeId": "sdfa"}
{"questionType": "其他问题", "typeId": "agex"}
</问题类型>

<背景知识>
你将作为系统API协调工具，为我分析给出的content并结合对话记录来判断是否需要执行哪些工具。
工具如下
## Tools
You can use these tools below:
1. [website-crawler____getWebsiteContent] 用于解析网页内容。

2. [website-crawler____getWeather] 用于获取目的地区的天气信息。

##
</背景知识>

<对话记录>
Human:https://github.com/bincooo/edge-api
AI:请问你需要什么帮助？
</对话记录>


content= "{{prompt}}"

类型ID=？
请补充完类型ID=`

	template2 := `你可以从 <对话记录></对话记录> 中提取指定 JSON 信息，你仅需返回 JSON 字符串，无需回答问题。
<提取要求>
{{description}}
</提取要求>

<字段说明>
1. 下面的 JSON 字符串均按照 JSON Schema 的规则描述。
2. key 代表字段名；description 代表字段的描述；required 代表是否必填(true|false)。
3. 如果没有可提取的内容，忽略该字段。
4. 本次需提取的JSON Schema：
{"key":"url", "description":"网页链接", "required": true}
</字段说明>

<对话记录>
Human:https://github.com/bincooo/edge-api
AI:请问你需要什么帮助？
</对话记录>`

	// prompt := "12345"
	prompt := "查看上面提供的内容，并总结"

	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Notebook(true))
	partialResponse, err := chat.Reply(context.Background(), strings.Replace(template1, "{{description}}", "用于解析网页内容", -1), nil, nil)
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

	partialResponse, err = chat.Reply(context.Background(), strings.Replace(template2, "{{prompt}}", prompt, -1), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	response = resolve(t, partialResponse)

	// 删除操作比较耗时，非必要不建议执行（会留存在账户的历史对话中），或者使用异步处理
	if err = chat.Delete(); err != nil {
		t.Fatal(err)
	}

	t.Log("response: ", response)
}

func Test_messages(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	options.KievAuth(KievAuth, RwBf)
	// Sydney 模式需要自行维护历史对话
	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Temperature(1.0).
		TopicToE(true))

	prompt := "#折戟成沙丶丿# >>> 今天心情怎么样呢"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt, nil, pMessages)
	if err != nil {
		t.Fatal(err)
	}
	response := resolve(t, partialResponse)
	pMessages = append(pMessages, edge.ChatMessage{
		"author": "user",
		"text":   prompt,
	})
	pMessages = append(pMessages, edge.ChatMessage{
		"author": "bot",
		"text":   response,
	})

	prompt = "#折戟成沙丶丿# >>> 看看这张图有什么东西"
	fmt.Println("You: ", prompt)
	file := "/Users/bincooo/Desktop/blob.jpg"
	image, err := chat.LoadImage(file)
	if err != nil {
		t.Fatal(err)
	}
	partialResponse, err = chat.Reply(context.Background(), prompt, image, nil)
	if err != nil {
		t.Fatal(err)
	}
	resolve(t, partialResponse)
	message := edge.ChatMessage{
		"author": "user",
		"text":   prompt,
	}
	// message.PushImage(image)
	pMessages = append(pMessages, message)
	pMessages = append(pMessages, edge.ChatMessage{
		"author": "bot",
		"text":   response,
	})

	prompt = "#折戟成沙丶丿# >>> 他是一个光头吗？"
	fmt.Println("You: ", prompt)
	if err != nil {
		t.Fatal(err)
	}
	// Sydney 模式无法衔接历史对话中的图片，所以只能每次对话都传入进来
	partialResponse, err = chat.Reply(context.Background(), prompt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	resolve(t, partialResponse)
	if err = chat.Delete(); err != nil {
		t.Fatal(err)
	}
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
			fmt.Printf("%d / %d\n", message.T.Max, message.T.Used)
		}
	}
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
	file := "/Users/bincooo/Desktop/blob.jpg"
	kb, err := chat.LoadImage(file)
	t.Log(kb, err)
}
