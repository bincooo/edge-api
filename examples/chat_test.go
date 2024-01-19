package main

import (
	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"testing"
)

const (
	cookie = "1VzW0-qcUhC-XBS6gZ-Y2hggKE6FX9ge8NI5GfbCXpO5vCrh8C5SZ6kKzUu7IZM1dCevyryxmK96Kl5kYVfAGl9m0Mcrmy8oThwPocaaXuAOUVpEDbF5AmJKrbwz0Ge3XCbtlKNwA24yfoMQYCXK85GMRCSTiHogXUA3unf7tFcFsszdWeBGSz1-J3OkWL1QzRJyS3YRJaMwjxu6CUw_4EQ"

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

func Test_messages(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "")
	if err != nil {
		t.Fatal(err)
	}
	//options.KievAuth(KievAuth, RwBf)
	options.Proxies("http://127.0.0.1:7890")
	// Sydney 模式需要自行维护历史对话
	options.Model(edge.ModelCreative)
	options.Temperature(1.0)
	options.TopicToE(true)
	chat := edge.New(options)

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

		msg += message.Text
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
