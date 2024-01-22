package edge

import (
	_ "embed"
	"github.com/joho/godotenv"
	"os"
)

const (
	//DefaultCreate  = "https://www.bing.com/turing/conversation/create"
	DefaultCreate  = "https://copilot.microsoft.com/turing/conversation/create"
	DefaultChatHub = "wss://sydney.bing.com/sydney/ChatHub"

	ModelCreative = "Creative"
	ModelBalanced = "Balanced"
	ModelPrecise  = "Precise"
	ModelSydney   = "Sydney"

	PluginShop      = "Shop"
	PluginInstacart = "Instacart"
	PluginOpenTable = "OpenTable"
	PluginKlarna    = "Klarna"
	PluginSearch    = "Search"
	PluginKayak     = "Kayak"

	delimiter = '\u001E'
)

//go:embed chat.json
var chatHub []byte

//go:embed notebook.json
var nbkHub []byte

var Version = "1.1482.6"

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

	schema = []byte{123, 34, 112, 114, 111, 116, 111, 99, 111, 108, 34, 58, 34, 106, 115, 111, 110, 34, 44, 34, 118, 101, 114, 115, 105, 111, 110, 34, 58, 49, 125, 30}
	ping   = []byte{123, 34, 116, 121, 112, 101, 34, 58, 54, 125, 30}
)

func init() {
	_ = godotenv.Load()
	Version = LoadEnvVar("BING_VER", Version)
}

func LoadEnvVar(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
