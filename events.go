package ranni

type EventType int

const (
	GroupMessageEventType EventType = iota
	PrivacyMessageEventType
)

func (event EventType) String() string {
	switch event {
	case GroupMessageEventType:
		return "group"
	case PrivacyMessageEventType:
		return "private"
	}
	panic("unknown eventType")
}

type Event interface {
	EventType() EventType
}

type BaseEvent struct {
	Time     int64  `json:"time"`
	SelfId   int64  `json:"self_id"`
	PostType string `json:"post_type"`
}

type MessageEvent struct {
	BaseEvent
	MessageType string `json:"message_type"`
	SubType     string `json:"sub_type"`
	MessageId   int32  `json:"message_id"`
	UserId      int64  `json:"user_id"`
	RawMessage  string `json:"raw_message"`
	Font        int32  `json:"font"`
	Sender      Sender `json:"sender"`
	MessageChain
}

// GroupMessageEvent 群聊消息事件
type GroupMessageEvent struct {
	MessageEvent
	GroupId   int64     `json:"group_id"`
	Anonymous Anonymous `json:"anonymous"`
}

func (groupMessageEvent GroupMessageEvent) EventType() EventType {
	return GroupMessageEventType
}

// PrivacyMessageEvent 私聊消息事件
type PrivacyMessageEvent struct {
	MessageEvent
}

func (privacyMessageEvent PrivacyMessageEvent) EventType() EventType {
	return PrivacyMessageEventType
}

// Anonymous 匿名
type Anonymous struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}

// Sender 发送者信息
type Sender struct {
	UserId   int64  `json:"user_id"`
	NickName string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}
