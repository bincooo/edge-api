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

	PluginInstacart = "www.instacart.com"
	PluginKlarna    = "www.klarna.com"
	PluginKayak     = "www.kayak.com"
	PluginShop      = "shop.app"
	PluginOpenTable = "www.opentable.com"
	PluginSearch    = "www.bing.com"
	PluginPhone     = "aka.ms"
	PluginSuno      = "www.suno.ai"

	delimiter = '\u001E'

	//

	OptionSets_Clgalileonsr = "clgalileonsr" // ???
	OptionSets_Nosearchall  = "nosearchall"  // 不联网查询
)

//go:embed chat.json
var chatHub []byte

//go:embed notebook.json
var nbkHub []byte

var Version = "1.1742.1"

var (
	schema = []byte{123, 34, 112, 114, 111, 116, 111, 99, 111, 108, 34, 58, 34, 106, 115, 111, 110, 34, 44, 34, 118, 101, 114, 115, 105, 111, 110, 34, 58, 49, 125, 30}
	ping   = []byte{123, 34, 116, 121, 112, 101, 34, 58, 54, 125, 30}
	end    = []byte{123, 34, 116, 121, 112, 101, 34, 58, 55, 125}
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
