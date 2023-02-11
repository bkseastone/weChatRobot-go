package controller

import (
	"fmt"
	"log"
	"net/http"
	"weChatRobot-go/models"
	"weChatRobot-go/service"

	"github.com/gin-gonic/gin"
)

func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

type MessageController struct {
	WechatService service.WechatService
}

// ReceiveMessage 收到微信回调信息
func (mc *MessageController) ReceiveMessage(c *gin.Context) {
	if c.Request.Method == "GET" {
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")
		if mc.WechatService.CheckSignature(signature, timestamp, nonce) {
			fmt.Println("chenggong")
			_, _ = fmt.Fprint(c.Writer, c.Query("echostr"))
		} else {
			fmt.Println("shibai")
			_, _ = fmt.Fprint(c.Writer, "你是谁？你想干嘛？")
		}
	} else {
		var reqMessage models.ReqMessage
		err := c.ShouldBindXML(&reqMessage)
		if err != nil {
			_, _ = fmt.Fprint(c.Writer, "系统处理消息异常")
			log.Printf("解析XML出错: %v\n", err)
			return
		}

		log.Printf("收到消息 %v\n", reqMessage)
		respXmlStr := service.GetGPT3ResponseMessage(reqMessage)
		log.Printf("响应消息 ||%v||\n", respXmlStr)
		_, _ = fmt.Fprint(c.Writer, respXmlStr)
	}
}
