package ranni

type Config struct {
	WsAddr       string `yaml:"ws_addr"`
	CallBackAddr string `yaml:"call_back_addr"`
	AccessToken  string `yaml:"access_token"`
	ApiAddr      string `yaml:"api_addr"` // API端口
}
