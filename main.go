// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

var bot *linebot.Client
var Instruction = `1. 即時路況查詢
指令格式:
即時路況
[起點]
[終點]

2. 最佳路徑查詢
指令格式:
最佳路徑
[起點]
[終點]
[交通模式(開車, 走路, 大眾運輸, 自行車)]

3. 預測高峰時段
指令格式:
預測高峰時段
[起點]
[終點]

4. 道路施工查詢
指令格式:
道路施工查詢
[縣市名稱]`

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	cb, err := webhook.ParseRequest(os.Getenv("ChannelSecret"), r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	InstructionErrorMsg := "指令格式錯誤，請重新輸入指令，支援指令格式為:\n" + Instruction
	for _, event := range cb.Events {
		log.Printf("Got event %v", event)
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			// Handle only on text message
			case webhook.TextMessageContent:
				handleTextMessage(bot, e.ReplyToken, message.Text)

			default:
				if _, err = bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage(InstructionErrorMsg)).Do(); err != nil {
					log.Print(err)
				}
				log.Printf("Unknown message: %v", message)
			}
		case webhook.FollowEvent:
			log.Printf("message: Got followed event")
		case webhook.PostbackEvent:
			data := e.Postback.Data
			log.Printf("Unknown message: Got postback: " + data)
		case webhook.BeaconEvent:
			log.Printf("Got beacon: " + e.Beacon.Hwid)
		}
	}
}

func handleTextMessage(bot *linebot.Client, replyToken string, text string) {
	lines := strings.Split(text, "\n")
	function := strings.TrimSpace(lines[0])

	switch function {
	case "指令":
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("支援指令如下:\n"+Instruction)).Do(); err != nil {
			log.Print(err)
		}
	case "即時路況":
		if len(lines) != 3 {
			if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("指令格式錯誤，請重新輸入指令，支援指令格式為:\n\n"+Instruction)).Do(); err != nil {
				log.Print(err)
			}
			return
		}
		origin := strings.TrimSpace(lines[1])
		destination := strings.TrimSpace(lines[2])
		TrafficCondition := getTrafficCondition(origin, destination)
		reply := fmt.Sprintf("起點: %s\n終點: %s\n\n%s", origin, destination, TrafficCondition)
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
			log.Print(err)
		}

	case "最佳路徑":
		if len(lines) != 4 {
			if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("指令格式錯誤，請重新輸入指令，支援指令格式為:\n\n"+Instruction)).Do(); err != nil {
				log.Print(err)
			}
			return
		}
		origin := strings.TrimSpace(lines[1])
		destination := strings.TrimSpace(lines[2])
		mode := strings.TrimSpace(lines[3])
		switch mode {
		case "開車":
			mode = "driving"
		case "走路":
			mode = "walking"
		case "大眾運輸":
			mode = "transit"
		case "自行車":
			mode = "bicycling"
		default:
			if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("交通模式錯誤，請輸入: 開車, 走路, 大眾運輸, 或 自行車")).Do(); err != nil {
				log.Print(err)
			}
			return
		}
		bestRoute := getBestRoute(origin, destination, mode)
		reply := fmt.Sprintf("起點: %s\n終點: %s\n%s", origin, destination, bestRoute)
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
			log.Print(err)
		}

	case "預測高峰時段":
		if len(lines) != 3 {
			if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("指令格式錯誤，請重新輸入指令，支援指令格式為:\n\n"+Instruction)).Do(); err != nil {
				log.Print(err)
			}
			return
		}
		origin := strings.TrimSpace(lines[1])
		destination := strings.TrimSpace(lines[2])
		reply := fmt.Sprintf(getPredictedTraffic(origin, destination))
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
			log.Print(err)
		}
	case "道路施工查詢":
		if len(lines) != 2 {
			if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("指令格式錯誤，請重新輸入指令，支援指令格式為:\n\n"+Instruction)).Do(); err != nil {
				log.Print(err)
			}
			return
		}
		target := strings.TrimSpace(lines[1])
		reply := GetConstruction(target)
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
			log.Print(err)
		}
	default:
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("指令格式錯誤，請重新輸入指令，支援指令格式為:\n\n"+Instruction)).Do(); err != nil {
			log.Print(err)
		}
	}
}
