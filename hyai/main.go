/*
 * Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"hy_golang_sdk/pkg/plog"
)

func synchronizeChat(client *TencentHyChat, query *HyQuery) string {
	newMsg := make([]Message, 0)

	for _, item := range query.History {
		newMsg = append(newMsg, item)
	}
	newMsg = append(newMsg, Message{
		Role:    "user",
		Content: query.Question,
	})
	resp, err := client.Chat(context.Background(), NewRequest(Synchronize, newMsg))
	if err != nil {
		plog.Errorf("tencent hunyuan chat err:%+v\n", err)
		return fmt.Sprintf("error:%v", err)
	}

	fmt.Printf("同步访问结果: 问题:%v", query.Question)
	for res := range resp {
		if res.Error.Code != 0 {
			plog.Errorf("tencent hunyuan chat err:%+v\n", res.Error)
			return fmt.Sprintf("code:%v, msg:%v", res.Error.Code, res.Error.Message)
		}
		//synchronize 同步打印message
		return res.Choices[0].Messages.Content
	}
	return fmt.Sprintf("error response:%v", "empty ")
}

// 流式访问
func streamChat(client *TencentHyChat, query *HyQuery) (<-chan Response, error) {
	newMsg := make([]Message, 0)

	for _, item := range query.History {
		newMsg = append(newMsg, item)
	}
	newMsg = append(newMsg, Message{
		Role:    "user",
		Content: query.Question,
	})
	resp, err := client.Chat(context.Background(), NewRequest(Stream, newMsg))
	if err != nil {
		fmt.Printf("tencent hunyuan chat err:%+v\n", err)
		return nil, err
	}
	fmt.Printf("流式访问结果: 问题：%v ", query.Question)
	return resp, nil
}

var hyClient *TencentHyChat

func main() {
	//{"history:[{"role":"user","content":"给我讲个冷笑话"}, {"role":"assistant", "content":"电脑为什么冷，因为它有window"}], "query":"不好笑"}
	//登陆控制台获取appID和密钥信息 替换下面的值
	var SecretID = ""
	var SecretKey = ""
	var appID int64 = 1310744949

	credential := NewCredential(SecretID, SecretKey)
	hyClient = NewTencentHyChat(appID, credential)
	plog.InitFileLogger(".", "logic")
	ginInstance := gin.Default()
	ossG := ginInstance.Group("hyai")
	ossG.POST("query", hyQuery)

	err := ginInstance.Run(":30033")
	if err != nil {
		return
	}
	//synchronizeChat(client) //同步访问
	//streamChat(client)      //流式访问

}

func hyQuery(ctx *gin.Context) {
	var param HyQuery
	if err := ctx.ShouldBind(&param); err != nil || len(param.Question) <= 0 {
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(400, gin.H{"error": "param empty"})
		}
		return
	}
	if !param.Stream {
		queryResult := synchronizeChat(hyClient, &param)
		ctx.String(200, queryResult)
	} else {
		resp, err := streamChat(hyClient, &param)
		if err != nil {
			ctx.String(500, "error:%v", err.Error())
			return
		}

		for res := range resp {
			if res.Error.Code != 0 {
				ctx.Writer.Write([]byte(fmt.Sprintf("err: (code:%v, msg:%v)", res.Error.Code, res.Error.Message)))
				fmt.Printf("tencent hunyuan chat err:%+v\n", res.Error)
				break
			}
			ctx.Writer.Write([]byte(fmt.Sprintf("%v", res.Choices[0].Delta.Content)))
			ctx.Writer.Flush()
			//stream  流式打印Delta
			fmt.Print(res.Choices[0].Delta.Content, "\n")
		}
	}

}

type HyQuery struct {
	History  []Message `json:"history" form:"history"`
	Question string    `json:"query" form:"query"`
	Stream   bool      `json:"stream" form:"stream"`
}

//curl -H "Content-type: application/json" -X POST -d  '{"stream", true, "history":[],"query":"不好笑"}' http://localhost:30030/hyai/query
