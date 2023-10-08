package main

import (
	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"io"
)

func main() {

	const (
		cookie = "xxx"
		agency = ""

		KievAuth = "xxx"
		RwBf     = "xxx"
	)
	chat, err := edge.New(cookie, agency)
	chat.KievRPSSecAuth = KievAuth
	chat.RwBf = RwBf
	chat.Proxy = "http://127.0.0.1:7890"
	chat.Model = edge.Sydney
	if err != nil {
		panic(err)
	}

	prompt := "11{blob:r83W0H-iQysG1g#rwtohAo3oCsGwA}\n这是什么游戏? 简单回答"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	pMessages := make([]map[string]string, 0)
	pMessages = append(pMessages, map[string]string{
		"author": "user",
		"text":   prompt,
	})

	prompt = "这张图显示了数字几？"
	fmt.Println("You: ", prompt)
	partialResponse, err = chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	pMessages = append(pMessages, map[string]string{
		"author": "bot",
		"text":   Println(partialResponse),
	})

	//prompt = "what can you do?"
	//fmt.Println("You: ", prompt)
	//timeout, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
	//defer cancel()
	//partialResponse, err = chat.Reply(timeout, prompt, nil)
	//if err != nil {
	//	panic(err)
	//}
	//Println(partialResponse)
}

func Println(partialResponse chan edge.PartialResponse) string {
	msg := ""
	for {
		message, ok := <-partialResponse
		if !ok {
			return msg
		}

		if message.Error != nil {
			if message.Error == io.EOF {
				return msg
			}
			panic(message.Error)
		}

		msg += message.Text
		fmt.Println(message.Text)
		fmt.Println("===============")
	}
}
