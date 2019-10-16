// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"encoding/json"
	"fmt"

	"github.com/laiye-ai/wulai-openapi-sdk-golang/services/common/errors"
	"github.com/laiye-ai/wulai-openapi-sdk-golang/services/common/log"
	"github.com/laiye-ai/wulai-openapi-sdk-golang/services/wulai"
)

//MsgBody 消息体
type MsgBody struct {
	cid  string //client id
	mid  string //message id
	data []byte //message data
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	//机器人客户端
	wulai *wulai.Client

	// Registered clients.
	clients map[*Client]bool

	//接受用户消息
	userMsgQueue chan *MsgBody

	//发送消息给用户
	echoMsgQueue chan *MsgBody

	//机器人消息
	botMsgQueue chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

//NewHub New Hub
func NewHub(pubkey, secret string) *Hub {
	hub := &Hub{
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		clients:      make(map[*Client]bool),
		userMsgQueue: make(chan *MsgBody, 100),
		echoMsgQueue: make(chan *MsgBody, 100),
		botMsgQueue:  make(chan []byte),
		wulai:        wulai.NewClient(secret, pubkey),
	}

	return hub
}

//Run start service
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case userMsg := <-h.userMsgQueue:
			//1:将用户消息发送给机器人
			h.echoMsgQueue <- doUserMsg(h, userMsg)
		case botMsg := <-h.botMsgQueue:
			//2:接受机器人回复的消息
			h.echoMsgQueue <- doMsgDelivery(h, botMsg)
		case echoMsg := <-h.echoMsgQueue:
			//3:将机器人回复发给用户
			for client := range h.clients {
				if echoMsg.cid != client.id {
					continue
				}
				select {
				case client.send <- echoMsg.data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

//doUserMsg 1:处理用户消息
func doUserMsg(h *Hub, userMsg *MsgBody) (botMsg *MsgBody) {

	botMsg = &MsgBody{userMsg.cid, userMsg.mid, []byte("发送失败")}
	log.Infof("[%s]发送消息给机器人:%v-%s\n", userMsg.cid, userMsg.mid, userMsg.data)

	//1创建用户
	user, err := h.wulai.UserCreate(userMsg.cid, userMsg.cid, "")
	if err != nil {
		log.Fatalf("user Create test reuslt:%s", err.Error())
		return botMsg
	}
	log.Infof("[创建用户成功]: %+v\n", user)

	//消息类型[文本消息]
	textMsg := &wulai.Text{
		Content: string(userMsg.data),
	}

	//2:发起问答
	respBody, err := h.wulai.MsgReceive(userMsg.cid, textMsg, fmt.Sprintf("%v", userMsg.mid), "预留信息")
	if err != nil {
		if cliErr, ok := err.(*errors.ClientError); ok {
			log.Info(cliErr.Error())
		} else if serErr, ok := err.(*errors.ServerError); ok {
			log.Info(serErr.Error())
		}
		return botMsg
	}

	botMsg.data = []byte(userMsg.cid + ":" + string(userMsg.data))
	log.Info(respBody.MsgID)

	//3:将选择的答案同步给平台

	return botMsg
}

//doMsgDelivery 2:处理收到的消息投递
func doMsgDelivery(h *Hub, botMsg []byte) (userMsg *MsgBody) {

	bot := &wulai.MessageDelivery{}
	if err := json.Unmarshal(botMsg, bot); err != nil {
		log.Infof("[bot msg]%s\n", err)
		return userMsg
	}

	userMsg = &MsgBody{bot.UserID, bot.MsgID, []byte("小Q: " + bot.MsgBody.Text.Content)}
	log.Infof("[%s]接受到机器人的答案:%s-%s\n", bot.UserID, bot.MsgID, bot.MsgType)

	return userMsg
}
