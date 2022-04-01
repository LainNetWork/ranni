package ranni

import (
	"bytes"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	urls "net/url"
	"time"
)

var client = &http.Client{Timeout: 30 * time.Second}

func PostJson(url string, body interface{}, respStruct interface{}) error {
	marshal, _ := json.Marshal(body)
	u, _ := urls.Parse(url)
	values := u.Query()
	values.Add("access_token", robotConfig.AccessToken)
	u.RawQuery = values.Encode()
	resp, err := client.Post(u.String(), "application/json", bytes.NewBuffer(marshal))
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		if respStruct == nil {
			return nil
		}
		all, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			return err2
		}
		err := json.Unmarshal(all, respStruct)
		if err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

func GetWithParams(url string, params urls.Values, respStruct interface{}, path ...interface{}) error {
	parse, err := urls.Parse(url)
	if err != nil {
		return err
	}
	params.Add("access_token", robotConfig.AccessToken)
	parse.RawQuery = params.Encode()
	urlWithParams := parse.String()
	resp, err := client.Get(urlWithParams)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		if respStruct == nil {
			return nil
		}
		all, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			return err2
		}
		json.Get(all, path...).ToVal(respStruct)
		return nil
	} else {
		return err
	}

}

type Result struct {
	IsOk bool        `json:"isOk"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NoAuth(ctx *gin.Context, msg string) {
	ctx.JSON(401, Result{
		IsOk: false,
		Msg:  msg,
	})
}

func Success(ctx *gin.Context, msg string) {
	ctx.JSON(200, Result{
		IsOk: true,
		Msg:  msg,
	})
}

func Error(ctx *gin.Context, msg string) {
	ctx.JSON(200, Result{
		IsOk: false,
		Msg:  msg,
	})
}

func SuccessWithData(ctx *gin.Context, data interface{}) {
	ctx.JSON(200, Result{
		IsOk: true,
		Data: data,
	})
}
