package config


type Config struct {
	
	Url string
	BotToken  string

	BasicAuth struct {
		USERNAME string
		PASSWORD string
	}

}

var Cfg Config