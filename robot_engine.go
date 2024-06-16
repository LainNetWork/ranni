package ranni

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron/v3"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var engine = &robotEngine{
	cronClient: cron.New(),
}

type robotEngine struct {
	HelpNotice     string
	innerListeners []EventHandler
	cronClient     *cron.Cron
}

func Start(config *Config) {
	engine.Start(config)
}

func Register(handler EventHandler) {
	engine.Register(handler)
}

func RegisterCron(cronStr string, cmd func()) {
	engine.RegisterCron(cronStr, cmd)
}

var robotConfig *Config

func (robotEngine *robotEngine) Start(config *Config) {
	robotConfig = config
	//启动定时器
	robotEngine.cronClient.Start()
	//启动web服务
	if robotConfig.ApiAddr != "" {
		go startApiServer()
	}
	//连接cq-http
	cqConnect()
}

func startApiServer() {
	e := gin.Default()
	e.POST("/send", ApiSendMessage)
	err := e.Run(robotConfig.ApiAddr)
	if err != nil {
		log.Println("api服务启动失败！", err)
	}
}

func cqConnect() {
	u := url.URL{Scheme: "ws", Host: robotConfig.WsAddr}
	values := url.Values{}
	values.Add("access_token", robotConfig.AccessToken)
	u.RawQuery = values.Encode()
	log.Printf("connecting to %s", robotConfig.WsAddr)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	client, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err.Error())
		return
	}
	defer func(client *websocket.Conn) {
		err := client.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(client)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := client.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			go func() {
				if jsoniter.Valid(message) {
					msgEvent := BaseEvent{}
					if err := jsoniter.Unmarshal(message, &msgEvent); err == nil {
						switch msgEvent.PostType {
						case "message":
							if event, err := messageEventDecode(message); err == nil {
								engine.CallEvent(event)
							}
						}
					}
				}
			}()
		}
	}()
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			err := client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func HelpNotice() string {
	return engine.HelpNotice
}

func (robotEngine *robotEngine) Register(listener EventHandler) {
	if len(robotEngine.HelpNotice) == 0 {
		robotEngine.HelpNotice = "使 用 指 南\n"
	}
	robotEngine.HelpNotice = robotEngine.HelpNotice + "\n" + listener.Help()
	if robotEngine.innerListeners == nil {
		robotEngine.innerListeners = make([]EventHandler, 0)
	}
	robotEngine.innerListeners = append(robotEngine.innerListeners, listener)
}

func (robotEngine *robotEngine) RegisterCron(cronStr string, cmd func()) {
	_, err := robotEngine.cronClient.AddFunc(cronStr, cmd)
	if err != nil {
		log.Println(err.Error())
		log.Panicln("定时任务添加异常")
	}
}

func (robotEngine *robotEngine) CallEvent(event Event) {
	context := &EventContext{}
	context.OriginalEvent = event
	context.Values = make(map[string]interface{})
	if event.EventType() == GroupMessageEventType {
		messageEvent := event.(GroupMessageEvent)
		context.GroupId = messageEvent.GroupId
		context.UserId = messageEvent.Sender.UserId
		context.EventType = GroupMessageEventType
		context.SelfId = messageEvent.SelfId
		context.Sender = messageEvent.Sender
		context.MessageChain = &messageEvent.MessageChain
	}
	if event.EventType() == PrivacyMessageEventType {
		messageEvent := event.(PrivacyMessageEvent)
		context.UserId = messageEvent.UserId
		context.EventType = PrivacyMessageEventType
		context.SelfId = messageEvent.SelfId
		context.Sender = messageEvent.Sender
		context.MessageChain = &messageEvent.MessageChain
	}
	for _, callBack := range robotEngine.innerListeners {
		go func(callBack EventHandler) {
			if callBack.Filter(context) {
				callBack.Do(context)
			}
		}(callBack)
	}
}

func messageEventDecode(post []byte) (event Event, err error) {
	messages := jsoniter.Get(post, "message")
	messageChain := JsonToMessageChain(messages)
	messageType := jsoniter.Get(post, "message_type").ToString()
	switch messageType {
	case "group":
		var p = &GroupMessageEvent{}
		err = jsoniter.Unmarshal(post, p)
		p.MessageChain = messageChain
		event = *p
	case "private":
		p := &PrivacyMessageEvent{}
		err = jsoniter.Unmarshal(post, p)
		p.MessageChain = messageChain
		event = *p
	default:
		return nil, errors.New("未知的消息类型！")
	}
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("解析事件内容错误！")
	}
	return event, nil
}

func JsonToMessageChain(messages jsoniter.Any) MessageChain {
	var msgS []Message
	for i := 0; i < messages.Size(); i++ {
		msgItem := messages.Get(i)
		data := msgItem.Get("data")
		switch msgItem.Get("type").ToString() {
		case Text.String():
			message := &TextMessage{}
			data.ToVal(message)
			msgS = append(msgS, *message)
		case Image.String():
			message := &ImageMessage{}
			data.ToVal(message)
			msgS = append(msgS, *message)
		case At.String():
			message := &AtMessage{}
			content := data.Get("qq").ToString()
			if content == "all" {
				message.AtAll = true
			} else {
				parseInt, err := strconv.ParseInt(content, 10, 64)
				if err != nil {
					log.Println("解析At对象QQ号失败！", err.Error())
					continue
				}
				message.Qq = parseInt
			}
			msgS = append(msgS, *message)
		case Reply.String():
			message := &ReplyMessage{}
			data.ToVal(message)
			msgS = append(msgS, *message)
		case Face.String():
			message := &FaceMessage{}
			data.ToVal(message)
			msgS = append(msgS, *message)
		case Record.String():
			message := &RecordMessage{}
			data.ToVal(message)
			msgS = append(msgS, *message)
		default:
			fmt.Println("无匹配类型", msgItem.ToString())
			continue
		}
	}
	return MessageChain{messages: msgS}
}
