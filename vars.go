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
		"901deletecos0",
		"emovoice",
		"kcinhero",
		"kcfullheroimg",
		"kcinlineels",
		"kcusenocutimg",
		"sydconfigoptt",
		"sydldfc",
		"0824cntor",
		"803iyjbexps0",
		"0529streamw",
		"streamw",
		"178gentechs0",
		"0825agicert",
		"804cdxedtgd",
		"0901usrprmpt",
		"019hlthgrds0",
		"829suggtrims0",
		"821fluxv13",
		"727nrprdrs0",
	}

	sSliceIds = []string{
		"222dtappid",
		"225cricinfo",
		"224locals0",
	}

	Schema = []byte("{\"protocol\":\"json\",\"version\":1}" + Delimiter)
	ping   = []byte("{\"type\":6}" + Delimiter)
)
