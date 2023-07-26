package main

import (
	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"io"
)

func main() {

	const (
		token  = "xxx"
		agency = "https://edge.zjcs666.icu"
	)
	chat, err := edge.New(token, agency)
	chat.Model = edge.Sydney
	if err != nil {
		panic(err)
	}

	prompt := "一天有几个时辰"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	prompt = "今年发什么了什么"
	fmt.Println("You: ", prompt)
	partialResponse, err = chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

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

func Println(partialResponse chan edge.PartialResponse) {
	for {
		message, ok := <-partialResponse
		if !ok {
			return
		}

		if message.Error != nil {
			if message.Error == io.EOF {
				return
			}
			panic(message.Error)
		}

		fmt.Println(message.Text)
		fmt.Println("===============")
	}
}
