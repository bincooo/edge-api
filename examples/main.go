package main

import (
	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"io"
)

func main() {

	const (
		//cookie = ""
		cookie = "1bYAoFadY_27a8B5UuUElKUBRPxJ8P3py0fDs5Jtub2IKUjHl_lCL4CQRpgVawy6ulPDZ00iOfQxw2twclpKQjVtLT4vhz43_F9HbsbMpfNdN0ytX8UMW-Fwr_QbH_Mda_LR_H2e9hmHytI0ZCd6E5B_n48uANmMF4BIjjtJwtdppDrb24smeW-s9mwSeClik3mJMY4PUoV8MRDbsE6FSGQ"
		agency = ""
		// agency   = "https://sokwith-nbing.hf.space"
		KievAuth = "FAB6BBRaTOJILtFsMkpLVWSG6AN6C/svRwNmAAAEgAAACM4gmL7lky8LOASKw7t6ecbIFskhodSNqUFpO0rD+im/mAUS+UR9EfGnnRHLDl7Ei3QShr+caejZXvKOVHW3Qno1zHM41HYl4kLxT03Ne7T+jElYESyZCIKFCjSvA77EdXXAtonY20GJ/s8nwVJZKQNBaYLaRw8jdf27fUiXALDOLQLilWa+fYq/4jYBwTNUN9u29bz8lcpQcclj04r3qdCvAZUXerB2sL9rktN9erUP357tKAwoSXcHtDYQG00dbVcN3hvpAqEwPaq/xBXQ/UM/oP6AoUd2rV5ROFlZc6kJtSXo9yba/fobzfiQj3F0yV8g3Z0YWzRMlxbK92GygBvYkyJn2aFr5vmG4RAid4EhHjnRgD+Qz4rzkav2RfhNQ3/J1PH7moiAOvkXLYhJ2ROI8WVhe/M44D5jUK2rB0HJKdjIS36FZysRQTgxVnq0s3IQtMnjElMMORTdK75HxK7lOnqrD8eAdCCsHsbmpB7w6mgYCI3GNUrEF2q5UUr5opnvlF2JwMHRgSq6a1JPEckZNceSpDmgSbto9DB5FDlTii0xe0o89HHw/ZDIQIqGcn4q7OtqP83eyhA40uD6AYeoQDlseNuHpKkb5Vj/zTAa9HgrKT2FVcitPqOrfvwSVeyoYRIiwgm9sE8UuZdp66PrqC/JV661rkQ3atJzcRLLOjvZfqzKqP7y1dQaShxKUafE9idB4jo1AfRUN203WxO1V/pZl749uo0OLug8nKR5ImpQvtNAndLMxsRrPt/vA/AZoq6VGe1D0L9Ln3ZcA1h2gKfzDjlw9e4JkOU5rEhmSJu76eo8fkFJt18d8jNKuvl8DOm/Ib5vi0e2RL81OsXIPKenSt6WKKiiQw6zybFsJq2B7c+7Jw9RHbr9CbiQEObcz5xfuChw2AEpVdrbCtnVw2vyzwIp0t/jMX/1nhCZCDq+bmbrP6ZULmItfoQXd75x1FRB/RFXv5crB0ONF4s7RMxw6owPwHMth0GxiG4CLV+81Oef9VJohMpDr3HXZZ36MzYoKmqkmrZPoTVlf4DpY0jbEnVpoA/+lcMC9lvUS4btrMO7YW7gvaDdwPU6+9BB0xel0XRR4KbprFuTU/025KKqZPcTVmqAYFWG3BFjzr/qeZPo6g/NnZ7EO/xfsbio1s+rrimpX+n617SgW+2a+W7yWr4E8nZySNCo5brjiNoVawiHZ2bJq88XFnJrCC72EVfJ3Qtiit1G9IlUiwdZ/IWKhijm4IWJy0i4fFPPMx4qbtQmikCTT0Ziz0UvBbSUvQe5Nyp6rGF77UQzNv1zX1NDxLMHrIogS1yAxwSubJrWfbLVXRGGvi/V5X29Wu5sUrJufS/GWfTpvD7qJjUmg9r9nh5yeqZid4wcRVwFRElIefhGo/KKdFuk+mJuU/tw49C1USaBk59tWFobHcTQd6TloSZN2BbWUyKWZto8wZEUABRtVOAabjxDkZYszBLV3p5EsyHm"
		RwBf     = "r=1&ilt=2&ihpd=0&ispd=2&rc=3258&rb=3258&gb=0&rg=0&pc=3258&mtu=0&rbb=0.0&g=0&cid=&clo=0&v=4&l=2023-10-05T07:00:00.0000000Z&lft=0001-01-01T00:00:00.0000000&aof=0&o=16&p=bingcopilotwaitlist&c=MY00IA&t=8731&s=2023-02-09T05:43:58.6175399+00:00&ts=2023-10-05T18:13:22.2675304+00:00&rwred=0&wls=2&wlb=0&lka=0&lkt=0&TH=&mta=0&e=DEXUb252ZC9SFK4xHrsOiXrviFzKvU4xTYFhd2fcf9zIYWoAHoZxCvabGCapPO9hHfELbWR8GYg-bXFEOg3W03lVS7W-BVEoFln5poMCU0o&A=2FB211DA38AD5D6E14EB5C69FFFFFFFF"
	)
	chat, err := edge.New(cookie, agency)
	//chat.KievRPSSecAuth = KievAuth
	//chat.RwBf = RwBf
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
