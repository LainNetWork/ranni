package ranni

type Config struct {
	MasterAccount int64  `yaml:"master_account"`
	WsAddr        string `yaml:"ws_addr"`
	CallBackAddr  string `yaml:"call_back_addr"`
	AccessToken   string `yaml:"access_token"`
}
