## 吾来Bot异步定制对话场景案例

实现步骤

1. 调用UserCreate接口，传入 user_id，创建用户
2. 调用MsgReceive接口，将用户的问题发给吾来 
3. 在消息路由中接受机器人返回的结果，并根据业务需要修改相应内容
4. 吾来平台会回调消息投递接口，将最终回复发给用户

整体案例实现架构图:
<img src="./static/async_talk.png" width="800" height="400">

