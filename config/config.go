package config

import (
	"log"
	"path/filepath"

	"github.com/spf13/viper"
)

const(
	BROWSER_TIME_FORMAT = "2006-01-02T15:04"
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

	config.SetDefault("server.database.driver", "sqlite")
	config.SetDefault("server.database.dsn", "indiescrobble.db")

	config.SetConfigType("yaml")
	config.SetConfigName(env)
	config.AddConfigPath("../config/")
	config.AddConfigPath("config/")

	err = config.ReadInConfig()

	if err != nil {
		log.Fatal("error on parsing configuration file")
	}

	if config.GetString("jwt.signKey") == ""{
		log.Fatal("You must set a JWT sign key (jwt.signKey in config yaml)")
	}


	config.BindEnv("server.port","PORT")
	

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
