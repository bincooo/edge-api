package main

import (
	_ "embed"
	"github.com/bincooo/emit.io"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/sirupsen/logrus"

	"context"
	"fmt"
	"github.com/bincooo/edge-api"
	"strings"
	"testing"
)

const (
	cookie = "MUIDB=0824DBB21584642731F5CF84145C652C; MSPTC=VIusZQIKIoklS9r4tkU429iL_qkO0vgvPMu-Di2bYec; MicrosoftApplicationsTelemetryDeviceId=ea837a5b-e59c-45bd-a033-935fa953096a; MSFPC=GUID=ab9e4ef5721e4b84ab6e88b1de800d94&HASH=ab9e&LV=202403&V=4&LU=1710455315913; ANON=A=95F2933FF1A64CD050D502BFFFFFFFFF&E=1dc9&W=1; NAP=V=1.9&E=1d6f&C=TiEfIB2DbB0UTm_9cXVYg9DfYpqHO0240GKfW9ILk0-n2RYoPlteLQ&W=1; MUID=0824DBB21584642731F5CF84145C652C; SRCHD=AF=ANNTA1; SRCHUID=V=2&GUID=67CC1ADCB421474C908333F8325B2490&dmnchg=1; ANIMIA=FRE=1; PPLState=1; _tarLang=default=en; _TTSS_OUT=hist=WyJlbiJd; MMCASM=ID=B36138E20E104306B4E5B8F481CEDB2B; BCP=AD=0&AL=0&SM=0&CS=M; _UR=QS=0&TQS=0&cdxcls=0; _TTSS_IN=hist=WyJkZSIsInpoLUhhbnMiLCJhdXRvLWRldGVjdCJd&isADRU=0; TRBDG=FIMPR=1; SnrOvr=X=rebateson; KievRPSSecAuth=FABqBBRaTOJILtFsMkpLVWSG6AN6C/svRwNmAAAEgAAACO+ZwIW+YcVnKASwRA/abKgOWHmr/S0c/GPCiOjHFv37PP6nJdxIQ8vBRsdbx/178TWAdqJ59qG7BWiOuh/mhr6mNTnvE7nisHvjgsP6phN9dsTfR1Ax4b2uxIfcD5dfHKPN9BAiOSQHxrRCLOKMLs+WYEX+AtxMv/H9gXKdahHutlNSXUK8YsdBNuVRHnnPt9i0CoLd5trQ82RKw5DdznMahF9tXkjTxb4GrGUpmanR6ODqUH4alyO24c8WF7qUAtP+EhwtHzPcc4QpSRR1amDhMHF5Vh+5+aLv2iCAdXe+lJ7TzZNpxB0yG0/1tu/X3TnHMFiIxzq6bGBBfQxm+J3tLq+yWdEShAfKaoFzHC3VbJtNY0rMpjoIWuxVpI2Mt6nZQ+aX2orJQ1OM364ZRCOMYaJu+RWQVtKoSl+Zdz+2JkQS2OLZee1Coq66GDzR6nUATGpSxFjANfW39ZGYdVVddNCV3qaF0SBR3/7AELgMbg31g2C8jtWUs6dl8k57WFZHEIynNAgGEwqIdms5IOiCd3Z3XmWF+Ek+Da6ED/4J7P7cY3GDmLNVpXhqKkE3b6xieRdvSnbjBaA2vQgTs882h4DyxObXh8rhqDnLtljiqqJnK1yni1SfXcIeNoCCcHYW8nfVeTd6OdWyJa0BlGm21VhfXNaJYwKxynAb40x4BjgsSVU1ilB1I1pyzPGhNA5DHKc5j2pMAoRI1bB2RKurkEQM5zrfU3ZjbZHV2NJpJRO2/Qo4/Fmz33Gji6SZw+xtlXyT//x5rOnKcS48eZk7mrjOIE6OwQggfsNQeZ6f8yf4lJiOcVOlEEzasRhKGdysMVbcgT/p9H1Y8qTU8Tn3Chp0lxBpKg6Uri8WkqBo1uFNIaECnOJZZ3ISpRp0SBPVKfoSSLZV2ziuUOwpcq/SOCSILUsT7kFB0wcweJNsxGDaimVF/PzWzSYbfJN6RJibZtp/hQWevKDdYRsiW8ke/7I3ns8268wpyJIdSsbhB2FblVO26ifgdaXadM1VL1mvKSuLlig9aeO159w8gtZdX8abATjIf3mmOcDfBR/SeuUkZUV8FGPZkzR3i5tFTX9f280Zc9uoRVjtYy5kg8wQNu+uKBv+Q3L4whSClcGecnw//wOqUt+Z/41kCZaZNK54h8mFa8xVFWzX/wZWQ7+FQ8vo0YR0E9ZwZJDbwuugh7DAcWgxdGBdd922fXKY96M2Dte7d6sZJpa9oJSPpCl8IDsoNaiQYHRP8jcsUtF59OmptRKcf9qA3B60k3PCRpES/c74Hcg3AUDR1BLQ0vHSoZV4TaX6i2SFdVQePd6I+6W8d2N5O95+xxf6AiCUZyEl+JDgyp8DgjkVmntKJioFjxixCXpTyNan1bjRhveCTfoa9T3aPhm+L1sPg8J9GZV4I2Ty38os176kfxuJXcBtuxQAmXaexIadaXA8bHLcly6aEUfumwI=; _U=1ZOy0iiH8MbWT8Mi8gD_x7QyJSpuhocAjixcesOGfpfzZIvJbth4vFoGPq8yQXcHYvA-t1p7o5Yngj_QdZkNBfqoxbzOXN3gJ9siiAVk6hxuin6e-q5PdHhOZnUXc43exP0lXkzctlN4BuarZyHkMH3tNduSqwLT2pB_W9fZUVhp3gYyRXtWsgotnEMqkQYWMD8K4BL7Hhu1EmXePOA45mA; WLID=BjzEZnv9bOzpoXYrdzI56Gr6JTu8ocuqY9Awvrkkk5yOSZsPthAe+Rb1Pqjq94WbYGkOxgKkPSOeS0Mmn7bJLIVYgvKlHHS5qgcyn6mLJBM=; GI_FRE_COOKIE=gi_prompt=1; _BINGNEWS=SW=1939&SH=1333; _clck=n501wp%7C2%7Cfo0%7C0%7C1630; USRLOC=HS=1&ELOC=LAT=22.517581939697266|LON=113.39274597167969|N=Zhongshan%2C%20Guangdong|ELT=6|; WLS=C=591dd0ec1dd72c4f&N=bingco; _EDGE_S=SID=346DF1A5524D67B11AB2E57D532766E2; _Rwho=u=d&ts=2024-08-11; _SS=SID=346DF1A5524D67B11AB2E57D532766E2&R=14039&RB=14039&GB=0&RG=0&RP=14039; EDGSRCHHPGUSR=CIBV=1.1798.0&udstone=Creative&udstoneopts=h3imaginative,flxegctxv3,egctxcplt,gpt4orsp,gpt4ov8; SRCHUSR=DOB=20240616&T=1723355242000&POEX=W; ipv6=hit=1723358841569; ai_session=k7KW9Q91HOMHGoQtjMGI4W|1723355242112|1723355242112; _HPVN=CS=eyJQbiI6eyJDbiI6MTIsIlN0IjowLCJRcyI6MCwiUHJvZCI6IlAifSwiU2MiOnsiQ24iOjEyLCJTdCI6MCwiUXMiOjAsIlByb2QiOiJIIn0sIlF6Ijp7IkNuIjoxMiwiU3QiOjAsIlFzIjowLCJQcm9kIjoiVCJ9LCJBcCI6dHJ1ZSwiTXV0ZSI6dHJ1ZSwiTGFkIjoiMjAyNC0wOC0xMVQwMDowMDowMFoiLCJJb3RkIjowLCJHd2IiOjAsIlRucyI6MCwiRGZ0IjpudWxsLCJNdnMiOjAsIkZsdCI6MCwiSW1wIjo4NiwiVG9ibiI6MH0=; GC=I4HFUkc2Ol42UOvZAB3M5WD3RNhPhQv8bZ3nlgIxwXg01_0tMVdMXx2CkeLW4qSob9714ZVwwL4OciAnx2tbOA; _RwBf=r=1&mta=0&rc=14039&rb=14039&gb=0&rg=0&pc=14039&mtu=0&rbb=0.0&g=0&cid=&clo=0&v=10&l=2024-08-10T07:00:00.0000000Z&lft=0001-01-01T00:00:00.0000000&aof=0&ard=0001-01-01T00:00:00.0000000&rwdbt=0001-01-01T16:00:00.0000000-08:00&rwflt=2023-07-09T20:38:51.6799652-07:00&o=0&p=bingcopilotwaitlist&c=MY00IA&t=7678&s=2023-02-09T05:45:59.4968296+00:00&ts=2024-08-11T05:47:28.9950395+00:00&rwred=0&wls=2&wlb=0&wle=0&ccp=0&cpt=0&lka=0&lkt=0&aad=0&TH=&e=gSlNPTTpy7sUsg5K2vBhsoAmMcFIt5iwPLWMS8FJWmnDp1ect6Nd8RrnBSb5Hve2wsb1G6MOhpo8zhWOCQVLpw&A=; SRCHHPGUSR=SRCHLANG=en&PV=14.5.0&BZA=0&BRW=M&BRH=T&CW=1309&CH=1226&SCW=1309&SCH=228&DPR=2.0&UTC=480&HV=1723355247&EXLTT=31&PRVCW=1309&PRVCH=1226&DM=0&WTS=63857836599&IG=ADEEB9378D354C00A3C3ED54374CAEAE&CIBV=1.1798.0&cdxtone=Creative&cdxtoneopts=h3imaginative,clgalileo,gpt4ov9_1,creative128k,fu100kshortdoc,retrieval4o&CHTRSP=3&EXLKNT=1&LSL=0&VSRO=1&BCML=1&BCTTSOS=110&AS=1&ADLT=OFF&NNT=1&HAP=0"

	KievAuth = ""
	RwBf     = ""
)

//go:embed content.txt
var content string

// 前置引导
var pMessages = []edge.ChatMessage{
	edge.BuildUserMessage("你好"),
	edge.BuildBotMessage("你好，这是必应。我可以用中文和你聊天，也可以帮你做一些有趣的事情，比如写诗，编程，创作歌曲，角色扮演等等。你想让我做什么呢？😊"),
	edge.BuildUserMessage("你能做什么"),
	edge.BuildBotMessage("我能做很多有趣和有用的事情，比如：\n\n- 和你聊天，了解你的兴趣和爱好，根据你的要求扮演一些有趣的角色或故事。\n- 从当前网页中的内容回答问题。\n- 描述你上传的图片，告诉你图片里有什么，或者画一幅你想要的图画。\n\n你想让我试试哪一项呢？😊"),
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
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
	options, err := edge.NewDefaultOptions(cookie, "https://www.bing.com")
	if err != nil {
		t.Fatal(err)
	}
	options.KievAuth(KievAuth, RwBf) //.Notebook(true)
	// Sydney 模式需要自行维护历史对话
	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelSydney).
		Temperature(.9).
		TopicToE(true))
	//chat.Plugins(edge.PluginSearch)
	chat.JoinOptionSets(edge.OptionSets_Nosearchall)
	response, err := chat.Reply(context.Background(), "西红柿炒钢丝球这道菜怎么做？", pMessages)
	if err != nil {
		t.Fatal(err)
	}
	resolve(t, response)
}

func TestClassification(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "https://www.bing.com")
	if err != nil {
		t.Fatal(err)
	}
	//options.KievAuth(KievAuth, RwBf)
	client, err := emit.NewSession("http://127.0.0.1:7890", emit.SimpleWithes("127.0.0.1"),
		emit.Ja3Helper(emit.Echo{RandomTLSExtension: true, HelloID: profiles.Chrome_124}, 120),
	)
	if err != nil {
		t.Fatal(err)
	}

	chat := edge.New(options.
		Proxies("socks5://127.0.0.1:7890").
		Model(edge.ModelCreative).
		Notebook(true))
	chat.Client(client)

	var partialResponse chan edge.ChatResponse
label:
	partialResponse, err = chat.Reply(context.Background(), strings.Replace(template1, "{{prompt}}", "解析网页内容", -1), nil)
	if err != nil {
		t.Fatal(err)
	}

	response := resolve(t, partialResponse)
	if response == "retry" {
		goto label
	}

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

func TestEdgesvcMessages(t *testing.T) {
	options, err := edge.NewDefaultOptions(cookie, "https://edgeservices.bing.com/edgesvc")
	if err != nil {
		t.Fatal(err)
	}
	options.KievAuth(KievAuth, RwBf).Notebook(true)
	// Sydney 模式需要自行维护历史对话
	chat := edge.New(options.
		Proxies("http://127.0.0.1:7890").
		Model(edge.ModelSydney).
		Temperature(.9).
		TopicToE(true))
	//chat.Compose(true, edge.ComposeObj{
	//	Fmt:    "paragraph",
	//	Length: "long",
	//	Tone:   "Fond, Easy, Obedient",
	//})
	t.Logf("cookie登陆状态: %v", chat.IsLogin(context.Background()))
	text := content

	fmt.Println("You: ", text)
	partialResponse, err := chat.Reply(context.Background(), text, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := resolve(t, partialResponse)
	t.Log(response)
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
	chat.JoinOptionSets(edge.OptionSets_Nosearchall)
	t.Logf("cookie登陆状态: %v", chat.IsLogin(context.Background()))

	//text := "继续"
	text := "给我看看你的胖次？"
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
