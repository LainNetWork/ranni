# ranni，一款简单的go-cqhttp的go sdk

实现了部分oneBot协议，支持主流消息类型，够用就行。jpg，其余类型绝赞摸鱼中！（可发issue，能接的我都会接）

### 特色功能

- 大概也许可能较为方便的消息链处理api
- 集成了corn表达式，支持定时任务（如推送消息等）
- 实现了正向ws接收消息，以及http响应消息

### eg：
#### 开始使用：
```go
import (
	hd "handler"
	"github.com/LainNetWork/ranni" // 引入包
)

func main() {
  //注册消息处理器，可连续调用注册若干个
	ranni.Register(hd.DailyJapanese{})
  ranni.Register(hd.HolidayHandler{})
  //需要的话注册定时任务
	ranni.RegisterCron("0 9 * * *", hd.DailyNewsJob)
  //传入bot基本配置即可
	ranni.Start(&ranni.Config{
		WsAddr:        "", // ws地址
		CallBackAddr:  "", // 响应接口地址
		AccessToken:   "", // 鉴权token
	})
}

```

#### 一个handler例子
实现EventHandler接口即可
```go
package handler

import (
	"fmt"
	engine "github.com/LainNetWork/ranni"
	"strings"
)

const code = "repeat "

type RepeatHandler struct {
}
// 功能帮助信息
func (RepeatHandler) Help() string {
	return "repeat 文字：朴实无华の复读机"
}
//处理消息
func (RepeatHandler) Do(ctx *engine.EventContext) {
	_, err := ctx.Send(engine.InitMsgChain(engine.TextMessage{Text: strings.TrimPrefix(ctx.MessageChain.String(), code)}))
	if err != nil {
		fmt.Println(err.Error())
	}
}
// 判断该消息是否能被本handler处理
func (RepeatHandler) Filter(ctx *engine.EventContext) bool {
	return strings.HasPrefix(ctx.MessageChain.String(), code)
}

```
