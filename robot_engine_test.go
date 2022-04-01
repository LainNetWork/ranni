package ranni

import (
	"testing"
	"time"
)

func Test_robotEngine_Start(t *testing.T) {
	Start(&Config{
		WsAddr:       "",
		CallBackAddr: "",
		AccessToken:  "",
		ApiAddr:      ":8999",
	})
	time.Sleep(100000 * time.Second)
}
