package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/laiye-ai/wulai-openapi-sdk-golang/services/common/log"
	"github.com/laiye-ai/wulai-openapi-sdk-golang/services/wulai"
)

// ServeMsgDelivery 消息投递 handles
func ServeMsgDelivery(hub *Hub, w http.ResponseWriter, r *http.Request) {
	//request log
	log.Infof("[/]=>remote=>%s host=>%s   url=>%s   method=>%s\n", r.RemoteAddr, r.Host, r.URL, r.Method)

	respBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Infof("%s\n", err)
		return
	}

	log.Infof("[机器人投递的消息]]=>%s\n", respBytes)
	//将消息投递到前端
	hub.botMsgQueue <- respBytes

	w.Write([]byte("ok"))
}

// ServeMsgRoute 消息路由 handles
func ServeMsgRoute(hub *Hub, w http.ResponseWriter, r *http.Request) {
	//request log
	log.Infof("[/]=>remote=>%s host=>%s   url=>%s   method=>%s\n", r.RemoteAddr, r.Host, r.URL, r.Method)

	inBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Infof("%s\n", err)
		return
	}

	log.Infof("[收到消息路由传递的消息]]=>%s\n", inBytes)

	//处理收到的消息
	msgBody := &wulai.MessageRoute{}
	if err := json.Unmarshal(inBytes, msgBody); err != nil {
		log.Errorf("%s\n", err)
	}

	respBody := &wulai.MessageRouteResponses{}
	respBody.IsDispatch = false                            //不转人工
	respBody.SuggestedResponse = msgBody.SuggestedResponse //不处理,直接将消息传回

	outBytes, _ := json.Marshal(respBody)

	log.Info("返回处理后的结果给机器人")
	w.Write(outBytes)
}
