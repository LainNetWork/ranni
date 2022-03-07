package ranni

import (
	"log"
	"strings"
)

type MessageType int

const (
	Text MessageType = iota
	Image
	Record
	Video
	At
	Reply
	Node
	Face
)

func (messageType MessageType) String() string {
	switch messageType {
	case Text:
		return "text"
	case Image:
		return "image"
	case Record:
		return "record"
	case Video:
		return "video"
	case At:
		return "at"
	case Reply:
		return "reply"
	case Node:
		return "node"
	case Face:
		return "face"
	default:
		log.Println("未知类型")
		return "unknown"
	}
}

type Message interface {
	MessageType() MessageType
	buildMessageMO() MessageMO
}

// TextMessage 文本消息
type TextMessage struct {
	Text string `json:"text"`
}

func (message TextMessage) buildMessageMO() MessageMO {
	return MessageMO{
		Type: message.MessageType().String(),
		Data: message,
	}
}

func (message TextMessage) MessageType() MessageType {
	return Text
}

type FaceMessage struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

func (message FaceMessage) buildMessageMO() MessageMO {
	//TODO implement me
	return MessageMO{
		Type: message.MessageType().String(),
		Data: message,
	}
}

func (message FaceMessage) MessageType() MessageType {
	return Face
}

// ImageMessage 图片消息
type ImageMessage struct {
	File string `json:"file"`
	Url  string `json:"url"`
	Type string `json:"type"` //为闪照时为flash
}

func (message ImageMessage) buildMessageMO() MessageMO {
	return MessageMO{
		Type: message.MessageType().String(),
		Data: message,
	}
}

func (message ImageMessage) MessageType() MessageType {
	return Image
}

// RecordMessage 语音消息
type RecordMessage struct {
	File  string `json:"file"`
	Magic string `json:"magic"` // 发送时可选，默认 0，设置为 1 表示变声
	Url   string `json:"url"`
}

func (message RecordMessage) buildMessageMO() MessageMO {
	return MessageMO{
		Type: message.MessageType().String(),
		Data: message,
	}
}

func (message RecordMessage) MessageType() MessageType {
	return Record
}

// VideoMessage 视频消息
type VideoMessage struct {
	File string `json:"file"`
	Url  string `json:"url"`
}

func (message VideoMessage) MessageType() MessageType {
	return Video
}

func (message VideoMessage) buildMessageMO() MessageMO {
	//TODO implement me
	return MessageMO{
		Type: message.MessageType().String(),
		Data: message,
	}
}

// AtMessage At消息
type AtMessage struct {
	Qq    int64 `json:"qq"`
	AtAll bool
}

func (message AtMessage) buildMessageMO() MessageMO {
	return MessageMO{
		Type: message.MessageType().String(),
		Data: message,
	}
}

func (message AtMessage) MessageType() MessageType {
	return At
}

type ReplyMessage struct {
	Id string `json:"id"`
}

func (message ReplyMessage) buildMessageMO() MessageMO {
	return MessageMO{
		Type: message.MessageType().String(),
		Data: message,
	}
}

func (message ReplyMessage) MessageType() MessageType {
	return Reply
}

type RedirectMessage struct {
	Name    string        `json:"name"`
	UserId  int64         `json:"user_id"`
	Content *MessageChain `json:"content"`
}

func (message RedirectMessage) buildMessageMO() MessageMO {
	return MessageMO{
		Type: message.MessageType().String(),
		Data: struct {
			Name    string       `json:"name"`
			UserId  int64        `json:"user_id"`
			Content *[]MessageMO `json:"content"`
		}{
			Name:    message.Name,
			UserId:  message.UserId,
			Content: buildMessageMO(message.Content),
		},
	}
}

func (message RedirectMessage) MessageType() MessageType {
	return Node
}

// MessageChain 消息链
type MessageChain struct {
	messages []Message
}

func (messageChain *MessageChain) Remove(position int) *MessageChain {
	if position < 0 || position >= len(messageChain.messages) {
		return messageChain
	}
	messages := messageChain.messages
	messageChain.messages = append(messages[:position], messages[position+1:]...)
	return messageChain
}

func (messageChain MessageChain) Split(num int) []*MessageChain {
	var result []*MessageChain
	messages := messageChain.messages
	count := len(messages) / num
	left := len(messages) % num
	for i := 0; i < count; i++ {
		i2 := messages[i*num : (i+1)*num]
		chain := MessageChain{i2}
		result = append(result, &chain)
	}
	if left != 0 {
		i := messages[len(messages)-left:]
		result = append(result, &MessageChain{i})
	}
	return result
}

func (messageChain MessageChain) GetMessages() []Message {
	return messageChain.messages
}

type MessageFilter func(message Message) bool

func (messageChain MessageChain) String() string {
	var msg []string
	for _, item := range messageChain.Match(Text).GetMessages() {
		message := item.(TextMessage)
		msg = append(msg, message.Text)
	}
	return strings.Trim(strings.Join(msg, ""), " ")
}

func (messageChain *MessageChain) Add(message Message) *MessageChain {
	messages := append(messageChain.GetMessages(), message)
	messageChain.messages = messages
	return messageChain
}

func (messageChain *MessageChain) AddText(content string) *MessageChain {
	messages := append(messageChain.GetMessages(), TextMessage{Text: content})
	messageChain.messages = messages
	return messageChain
}

func (messageChain *MessageChain) AddImage(url string) *MessageChain {
	messages := append(messageChain.GetMessages(), ImageMessage{File: url})
	messageChain.messages = messages
	return messageChain
}
func (messageChain *MessageChain) AppendChain(chain *MessageChain) *MessageChain {
	for _, message := range chain.messages {
		messageChain.Add(message)
	}
	return messageChain
}
func (messageChain *MessageChain) AddRecord(url string) *MessageChain {
	messages := append(messageChain.GetMessages(), RecordMessage{File: url})
	messageChain.messages = messages
	return messageChain
}

func (messageChain *MessageChain) AddAt(qq int64) *MessageChain {
	messages := append(messageChain.GetMessages(), AtMessage{Qq: qq})
	messageChain.messages = messages
	return messageChain
}

// Match 在消息链中匹配指定类型的消息
func (messageChain MessageChain) Match(msgType MessageType) MessageChain {
	var messages []Message
	for _, item := range messageChain.messages {
		if item.MessageType() == msgType {
			messages = append(messages, item)
		}
	}
	return MessageChain{
		messages: messages,
	}
}

func (messageChain MessageChain) ContainsAt(qq int64) bool {
	return messageChain.Match(At).Filter(func(message Message) bool {
		atMessage := message.(AtMessage)
		return atMessage.Qq == qq
	}).Exist()
}

func (messageChain MessageChain) Count() int {
	return len(messageChain.messages)
}

// Exist 在消息链中存在消息
func (messageChain MessageChain) Exist() bool {
	return messageChain.messages != nil && len(messageChain.messages) > 0
}

// Filter 在消息链中匹配符合过滤条件的消息
func (messageChain MessageChain) Filter(filter MessageFilter) MessageChain {
	var messages []Message
	for _, item := range messageChain.GetMessages() {
		if filter(item) {
			messages = append(messages, item)
		}
	}
	return MessageChain{messages: messages}
}
func (messageChain MessageChain) AnyMatch(filter MessageFilter) bool {
	for _, item := range messageChain.GetMessages() {
		if filter(item) {
			return true
		}
	}
	return false
}

func (messageChain MessageChain) FindFirst() Message {
	i := len(messageChain.GetMessages())
	if i <= 0 {
		return nil
	}
	return messageChain.GetMessages()[0]
}

func (messageChain MessageChain) FindLast() Message {
	i := len(messageChain.GetMessages())
	if i <= 0 {
		return nil
	}
	return messageChain.GetMessages()[i-1]
}

func NewMsgChain() *MessageChain {
	return &MessageChain{}
}

func InitMsgChain(message Message) *MessageChain {
	return &MessageChain{
		messages: []Message{message},
	}
}
