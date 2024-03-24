package ranni

import (
	"fmt"
	json "github.com/json-iterator/go"
	"log"
	"net/url"
	"strconv"
	"time"
)

// api列表
const (
	SendMessage         = "/send_msg"               //发送消息
	DeleteMessage       = "/delete_msg"             //撤回消息
	GetMessage          = "/get_msg"                //获取消息
	GetGroupMemberList  = "/get_group_member_list"  //获取群组人员列表
	SendGroupForwardMsg = "/send_group_forward_msg" //发送自定义合并转发消息
	GetLoginInfo        = "/get_login_info"
)

type EventContext struct {
	EventType     EventType     //group or private
	SelfId        int64         //robot的QQ号
	UserId        int64         //发送人QQ
	GroupId       int64         //群号
	Sender        Sender        //发送人详细信息
	MessageChain  *MessageChain //消息链
	OriginalEvent Event
	Values        map[string]interface{} //携带的参数
}

// GetSubjectId 获取聊天主题Id
func (event *EventContext) GetSubjectId() int64 {
	if event.EventType == GroupMessageEventType {
		return event.GroupId
	} else if event.EventType == PrivacyMessageEventType {
		return event.UserId
	} else {
		return -1
	}
}

type MessageCallBack struct {
	Data struct {
		MessageId int32 `json:"message_id"`
	} `json:"data"`
	RetCode int    `json:"retcode"`
	Status  string `json:"status"`
	Wording string `json:"wording"`
}

func FetchAvatarUrl(qq int64) string {
	return fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=640", qq)
}

// ReCall 撤回消息
func (messageCallBack MessageCallBack) ReCall(sec int) {
	if messageCallBack.Status == "failed" {
		return
	}
	go func() {
		time.Sleep(time.Duration(sec) * time.Second)
		var u = robotConfig.CallBackAddr + DeleteMessage
		err := PostJson(u, messageCallBack.Data, nil)
		if err != nil {
			log.Println(err.Error())
		}
	}()
}

type MessageMO struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type SendMessageMO struct {
	MessageType string      `json:"message_type"`
	UserId      int64       `json:"user_id"`
	GroupId     int64       `json:"group_id"`
	Message     []MessageMO `json:"message"`
}

func (event *EventContext) Send(message *MessageChain) (*MessageCallBack, error) {
	return send(event.EventType, event.GetSubjectId(), message)
}

func send(eventType EventType, id int64, message *MessageChain) (*MessageCallBack, error) {
	mo := buildMessageMO(message)
	var msgMO = SendMessageMO{
		MessageType: eventType.String(),
		UserId:      id,
		GroupId:     id,
		Message:     *mo,
	}
	var u = robotConfig.CallBackAddr + SendMessage
	back := &MessageCallBack{}
	err := PostJson(u, msgMO, back)
	if err != nil {
		return nil, err
	}
	return back, nil
}

func buildMessageMO(message *MessageChain) *[]MessageMO {
	var msgData []MessageMO
	for _, item := range message.GetMessages() {
		msgData = append(msgData, item.buildMessageMO())
	}
	return &msgData
}

type MessageReq struct {
	MessageId string `json:"message_id"`
}

func (event *EventContext) GetMessage(messageId string) (messageChain MessageChain, err error) {
	message := new(interface{})
	_ = PostJson(robotConfig.CallBackAddr+GetMessage, MessageReq{
		MessageId: messageId,
	}, &message)
	toString, err := json.MarshalToString(message)
	get := json.Get([]byte(toString), "data", "message")
	return JsonToMessageChain(get), nil
}

type GroupMemberList struct {
	Data    []GroupMemberData `json:"data"`
	RetCode int               `json:"retcode"`
	Status  string            `json:"status"`
}
type GroupMemberData struct {
	Age             int    `json:"age"`
	Area            string `json:"area"`
	Card            string `json:"card"`
	CardChangeable  bool   `json:"card_changeable"`
	GroupID         int    `json:"group_id"`
	JoinTime        int    `json:"join_time"`
	LastSentTime    int    `json:"last_sent_time"`
	Level           string `json:"level"`
	Nickname        string `json:"nickname"`
	Role            string `json:"role"`
	Sex             string `json:"sex"`
	ShutUpTimestamp int    `json:"shut_up_timestamp"`
	Title           string `json:"title"`
	TitleExpireTime int    `json:"title_expire_time"`
	Unfriendly      bool   `json:"unfriendly"`
	UserID          int64  `json:"user_id"`
}

func (event *EventContext) FetchGroupMemberList() (*GroupMemberList, error) {
	values := url.Values{}
	values.Set("group_id", strconv.FormatInt(event.GroupId, 10))
	resp := &GroupMemberList{}
	err := GetWithParams(robotConfig.CallBackAddr+GetGroupMemberList, values, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type BotInfo struct {
	Nickname string `json:"nickname"`
	UserId   int64  `json:"user_id"`
}

type BotInfoMO struct {
	Data    BotInfo `json:"data"`
	RetCode int     `json:"retcode"`
	Status  string  `json:"status"`
}

func GetBotInfo() *BotInfo {
	values := url.Values{}
	resp := &BotInfoMO{}
	err := GetWithParams(robotConfig.CallBackAddr+GetLoginInfo, values, resp)
	if err != nil {
		log.Println("获取bot信息异常", err.Error())
		return nil
	}
	return &resp.Data
}

func SendToGroup(id int64, message *MessageChain) (*MessageCallBack, error) {
	return send(GroupMessageEventType, id, message)
}

func SendForwardMsgToGroup(groupId int64, chain *MessageChain) (*MessageCallBack, error) {
	mo := struct {
		GroupId  int64       `json:"group_id"`
		Messages []MessageMO `json:"messages"`
	}{
		GroupId:  groupId,
		Messages: *buildMessageMO(chain),
	}
	back := &MessageCallBack{}
	err := PostJson(robotConfig.CallBackAddr+SendGroupForwardMsg, mo, back)
	if err != nil {
		return nil, err
	}
	return back, nil
}

func SendToPrivacy(id int64, message *MessageChain) (*MessageCallBack, error) {
	return send(PrivacyMessageEventType, id, message)
}
