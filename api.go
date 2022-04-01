package ranni

import "github.com/gin-gonic/gin"

// ApiMessageVO 只支持Text、Image、At三种格式
type ApiMessageVO struct {
	Type string `json:"type" binding:"required"`
	Url  string `json:"url"`
	Qq   int64  `json:"qq"`
	Text string `json:"text"`
}

func (apiMessageVO ApiMessageVO) ToMessage() Message {
	switch apiMessageVO.Type {
	case "Text":
		if apiMessageVO.Text == "" {
			return nil
		}
		return TextMessage{Text: apiMessageVO.Text}
	case "Image":
		if apiMessageVO.Url == "" {
			return nil
		}
		return ImageMessage{
			File: apiMessageVO.Url,
		}
	case "At":
		if apiMessageVO.Qq == 0 {
			return nil
		}
		return AtMessage{
			Qq:    apiMessageVO.Qq,
			AtAll: false,
		}
	default:
		return nil
	}
}

type MessageContent struct {
	Type     string         `json:"type" binding:"required"`     // privacy group
	Number   int64          `json:"number" binding:"required"`   //QQ号/群号
	Messages []ApiMessageVO `json:"messages" binding:"required"` //消息内容
}

func ApiSendMessage(ctx *gin.Context) {
	body := &MessageContent{}
	err := ctx.ShouldBindJSON(body)
	if err != nil {
		Error(ctx, "参数异常")
		return
	}
	chain := NewMsgChain()
	for _, message := range body.Messages {
		toMessage := message.ToMessage()
		if toMessage != nil {
			chain.Add(toMessage)
		}
	}
	switch body.Type {
	case "privacy":
		_, err := SendToPrivacy(body.Number, chain)
		if err != nil {
			Error(ctx, "发送失败！")
		}
	case "group":
		_, err := SendToGroup(body.Number, chain)
		if err != nil {
			Error(ctx, "发送失败！")
			return
		}
	}
}
