package server

import (
	"fmt"

	"git.jamesravey.me/ravenscroftj/indiescrobble/config"
)

func Init() {
	config := config.GetConfig()
	r := NewRouter()
	r.LoadHTMLGlob("templates/*.tmpl")
	r.Run( fmt.Sprintf("%v:%v", config.GetString("server.host"), config.GetString("server.port")))
}
