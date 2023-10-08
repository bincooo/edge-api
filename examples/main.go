package main

import (
	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"io"
)

func main() {

	const (
		cookie = "1jlu__K_COr23peG_PZfavtwnN0I6zZGzYbvcAwzeRtJ9W3cM9Fhhz2fLdrroIgZboR1XHt_HxCMClYd2fyZ41avh2zeQ3NWgBkCaKLtuxYH770bk4pPMi2ZJ-1aqB2KvPTohkC4orNnWB6kboIEJtrWOlE62cmp33CGtT9HAzGXI3M1hdAostItd881sjZpkwcECzcWYUSqnwe2qSvZyzQ"
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

	prompt := "{image:/Users/bincooo/Desktop/R-C.jpg}\n这是什么花"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	//prompt = "今年发什么了什么"
	//fmt.Println("You: ", prompt)
	//partialResponse, err = chat.Reply(context.Background(), prompt, nil)
	//if err != nil {
	//	panic(err)
	//}
	//Println(partialResponse)

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
