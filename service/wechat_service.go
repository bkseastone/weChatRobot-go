package service

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"log"
	"sort"
	"strings"
	"time"
	"weChatRobot-go/models"
)

type WechatService struct {
	Config models.WechatConfig
}

// CheckSignature 校验签名
func (ws *WechatService) CheckSignature(signature, timestamp, nonce string) bool {
	if signature == "" || timestamp == "" || nonce == "" {
		return false
	}

	arr := []string{ws.Config.Token, timestamp, nonce}
	// 将token、timestamp、nonce三个参数进行字典序排序
	sort.Strings(arr)
	//拼接字符串
	content := strings.Join(arr, "")
	//sha1签名
	sha := sha1.New()
	sha.Write([]byte(content))
	sha1Value := hex.EncodeToString(sha.Sum(nil))

	return signature == sha1Value
}

// GPT3
func GetGPT3ResponseMessage(reqMessage models.ReqMessage) string {
	var respMessage interface{}
	if reqMessage.MsgType == models.MsgTypeText {
		// resp, err := http.Get("http://localhost:8081/weChat/receiveMessage?query=" + reqMessage.Content)
		// if err != nil {
		// 	log.Fatalf("GET请求失败: %s", err)
		// }
		// defer resp.Body.Close()

		// body, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	log.Fatalf("读取响应失败: %s", err)
		// }
		// respMessage = BuildRespTextMessage(reqMessage.ToUserName, reqMessage.FromUserName, string(body))
		respMessage = BuildRespTextMessage(reqMessage.ToUserName, reqMessage.FromUserName, reqMessage.Content)
	} else {
		respMessage = BuildRespTextMessage(reqMessage.ToUserName, reqMessage.FromUserName, "我只对文字感兴趣[悠闲]")
	}

	if respMessage == nil {
		return ""
	} else {
		respXmlStr, err := xml.Marshal(&respMessage)
		if err != nil {
			log.Printf("XML编码出错: %v\n", err)
			return ""
		}

		return string(respXmlStr)
	}
}

// 消息回响
func GetEchoResponseMessage(reqMessage models.ReqMessage) string {
	var respMessage interface{}
	if reqMessage.MsgType == models.MsgTypeText {
		respMessage = BuildRespTextMessage(reqMessage.ToUserName, reqMessage.FromUserName, reqMessage.Content)
	} else {
		respMessage = BuildRespTextMessage(reqMessage.ToUserName, reqMessage.FromUserName, "我只对文字感兴趣[悠闲]")
	}

	if respMessage == nil {
		return ""
	} else {
		respXmlStr, err := xml.Marshal(&respMessage)
		if err != nil {
			log.Printf("XML编码出错: %v\n", err)
			return ""
		}

		return string(respXmlStr)
	}
}

func GetResponseMessage(reqMessage models.ReqMessage) string {
	var respMessage interface{}
	if reqMessage.MsgType == models.MsgTypeEvent {
		respMessage = GetRespMessageByEvent(reqMessage.ToUserName, reqMessage.FromUserName, reqMessage.Event)
	} else if reqMessage.MsgType == models.MsgTypeText {
		respMessage = GetRespMessageByKeyword(reqMessage.ToUserName, reqMessage.FromUserName, reqMessage.Content)
		if respMessage == nil {
			respMessage = GetRespMessageFromTuling(reqMessage.ToUserName, reqMessage.FromUserName, reqMessage.Content)
		}
	} else {
		respMessage = BuildRespTextMessage(reqMessage.ToUserName, reqMessage.FromUserName, "我只对文字感兴趣[悠闲]")
	}

	if respMessage == nil {
		return ""
	} else {
		respXmlStr, err := xml.Marshal(&respMessage)
		if err != nil {
			log.Printf("XML编码出错: %v\n", err)
			return ""
		}

		return string(respXmlStr)
	}
}

func GetRespMessageByEvent(fromUserName, toUserName, event string) interface{} {
	if event == models.EventTypeSubscribe {
		return BuildRespTextMessage(fromUserName, toUserName, "谢谢关注！可以开始跟我聊天啦😁")
	} else if event == models.EventTypeUnsubscribe {
		log.Printf("用户[%v]取消了订阅", fromUserName)
	}
	return nil
}

func GetRespMessageByKeyword(fromUserName, toUserName, keyword string) interface{} {
	v, ok := keywordMessageMap[keyword]
	if ok {
		msgType, err := v.Get("type").String()
		if err != nil {
			return nil
		}

		if msgType == models.MsgTypeText {
			content, _ := v.Get("Content").String()
			return BuildRespTextMessage(fromUserName, toUserName, content)
		} else if msgType == models.MsgTypeNews {
			articleArray, err := v.Get("Articles").Array()
			if err != nil {
				return nil
			}

			var articleLength = len(articleArray)
			var articles = make([]models.ArticleItem, articleLength)
			for i, articleJson := range articleArray {
				if eachArticle, ok := articleJson.(map[string]interface{}); ok {
					var article models.Article
					article.Title = eachArticle["Title"].(string)
					article.Description = eachArticle["Description"].(string)
					article.PicUrl = eachArticle["PicUrl"].(string)
					article.Url = eachArticle["Url"].(string)

					var articleItem models.ArticleItem
					articleItem.Article = article
					articles[i] = articleItem
				}
			}
			return BuildRespNewsMessage(fromUserName, toUserName, articles)
		}
	}
	return nil
}

func BuildRespTextMessage(fromUserName, toUserName, content string) models.RespTextMessage {
	respMessage := models.RespTextMessage{
		Content: models.CDATA{Text: content},
	}
	respMessage.FromUserName = models.CDATA{Text: fromUserName}
	respMessage.ToUserName = models.CDATA{Text: toUserName}
	respMessage.CreateTime = time.Now().Unix()
	respMessage.MsgType = models.CDATA{Text: "text"}
	return respMessage
}

func BuildRespNewsMessage(fromUserName, toUserName string, articles []models.ArticleItem) models.RespNewsMessage {
	respMessage := models.RespNewsMessage{
		ArticleCount: len(articles),
		Articles:     articles,
	}
	respMessage.FromUserName = models.CDATA{Text: fromUserName}
	respMessage.ToUserName = models.CDATA{Text: toUserName}
	respMessage.CreateTime = time.Now().Unix()
	respMessage.MsgType = models.CDATA{Text: "news"}
	return respMessage
}
