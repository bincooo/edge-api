package edge

import _ "embed"

const (
	DefaultCreate  = "https://www.bing.com/turing/conversation/create"
	DefaultChatHub = "wss://sydney.bing.com/sydney/ChatHub"

	Creative = "Creative"
	Balanced = "Balanced"
	Precise  = "Precise"
	Sydney   = "Sydney"

	Delimiter = "\u001E"
)

//go:embed chat.json
var chatHub []byte

var (
	sliceIds = []string{
		"winmuid3tf",
		"osbsdusgreccf",
		"ttstmout",
		"crchatrev",
		"winlongmsgtf",
		"ctrlworkpay",
		"norespwtf",
		"tempcacheread",
		"temptacache",
		"505scss0",
		"508jbcars0",
		"515enbotdets0",
		"5082tsports",
		"515vaoprvs",
		"424dagslnv1s0",
		"kcimgattcf",
		"427startpms0",
	}

	sSliceIds = []string{
		"222dtappid",
		"225cricinfo",
		"224locals0",
	}

	Schema = []byte("{\"protocol\":\"json\",\"version\":1}" + Delimiter)
	ping   = []byte("{\"type\":6}" + Delimiter)
)
