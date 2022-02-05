package config

import (
	"log"
	"path/filepath"

	"github.com/spf13/viper"
)

var config *viper.Viper

// Init is an exported method that takes the environment starts the viper
// (external lib) and returns the configuration struct.
func Init(env string) {
	var err error
	config = viper.New()

	config.SetDefault("indieauth.clientName", "https://indiescrobble.club")
	config.SetDefault("indieauth.redirectURL", "http://localhost:3000/auth")
	config.SetDefault("indieauth.oauthSubject", "IndieScrobble OAuth Client")
	config.SetDefault("indieauth.oauthCookieName","indiescrobble-oauth")
	config.SetDefault("indieauth.sessionSubject", "IndieScrobble Session")

	config.SetConfigType("yaml")
	config.SetConfigName(env)
	config.AddConfigPath("../config/")
	config.AddConfigPath("config/")

	err = config.ReadInConfig()

	if config.GetString("jwt.signKey") == ""{
		log.Fatal("You must set a JWT sign key (jwt.signKey in config yaml)")
	}


	config.BindEnv("server.port","PORT")
	
	if err != nil {
		log.Fatal("error on parsing configuration file")
	}
}

func relativePath(basedir string, path *string) {
	p := *path
	if len(p) > 0 && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}

func GetConfig() *viper.Viper {
	return config
}
