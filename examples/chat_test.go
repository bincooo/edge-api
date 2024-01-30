package main

import (
	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"strings"
	"testing"
)

const (
	cookie = "1fUHC2Yl9KH9ys47oNnIuGCFwzE-HlIb-NFmccmMfzroPu2vd3v1UbpoetLjGxAfGuNmaN5v7BgKXheHoZgLJx09N6JY4mDu0HFt9NFGIKaRipEsD1nxHDj8mB5nYDKYc91_XDOq38rx2glrHY0n_8f6Q-VcCnSv2bcePoY2sxLDsHZ0GrXzLxtrAdM4qWrV8ZBa_qdaioBP1eL62u1gRZg"

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
		Proxies("http://127.0.0.1:7890").
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
	template := `我会给你几个问题类型，请参考背景知识（可能为空）和对话记录，判断我“本次问题”的类型，并返回一个问题“类型ID”和“参数JSON”:
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
			parameters:
				url: {
					type: String
					description: 网页链接
				}
		2. [website-crawler____getWeather] 用于获取目的地区的天气信息。
			parameters:
				city: {
					type: String
					description: 城市名称
				}
		##
		
		不要访问contnet中的链接内容
		不可回复任何提示
		不允许做任何解释
		不可联网检索
		</背景知识>
		
		<对话记录>
		Human:https://github.com/bincooo/edge-api
		AI:请问你需要什么帮助？
		</对话记录>
		
		
		content= "{{prompt}}"
		
		类型ID=
		参数JSON=
		---
		补充类型ID以及参数JSON的内容。仅回复ID和JSON，不需要解释任何结果！
	`

	// prompt := "12345"
	prompt := "查看上面提供的内容，并总结"

	chat := edge.New(options.
		Proxies("http://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Notebook(true))
	partialResponse, err := chat.Reply(context.Background(), strings.Replace(template, "{{prompt}}", prompt, -1), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	response := resolve(t, partialResponse)
	t.Log("response: ", response)
	if strings.Contains(response, "wqre") {
		t.Log("assert: website-crawler____getWebsiteContent")
		left := strings.Index(response, "{")
		right := strings.LastIndex(response, "}")
		t.Log("args: ", response[left:right+1])
	} else if strings.Contains(response, "sdfa") {
		t.Log("assert: website-crawler____getWeather")
		left := strings.Index(response, "{")
		right := strings.LastIndex(response, "}")
		t.Log("args: ", response[left:right+1])
	} else {
		t.Log("assert: other")
	}

	// 删除操作比较耗时，非必要不建议执行（会留存在账户的历史对话中），或者使用异步处理
	//if err = chat.Delete(); err != nil {
	//	t.Fatal(err)
	//}
}

func Test_messages(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "https://bincooo-single-proxy.hf.space/copilot")
	if err != nil {
		t.Fatal(err)
	}
	//options.KievAuth(KievAuth, RwBf)
	// Sydney 模式需要自行维护历史对话
	chat := edge.New(options.
		Proxies("http://127.0.0.1:7890").
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
	file := "/Users/bincooo/Desktop/36452456.jpg"
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
	options.Proxies("http://127.0.0.1:7890")
	options.Model(edge.ModelSydney)
	chat := edge.New(options)
	file := "/Users/bincooo/Desktop/Screenshot 2023-11-11 at 11.21.23.png"
	kb, err := chat.LoadImage(file)
	t.Log(kb, err)
}
