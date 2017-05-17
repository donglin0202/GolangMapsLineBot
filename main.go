﻿// Licensed under the Apache License, Version 2.0 (the "License");
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

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Print("TextMessage: Type(" + message.Type + "), Text(" + message.Text  + ")" )
				//if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.ID+":"+message.Text+" OK!")).Do(); err != nil {
				//	log.Print(err)
				//}
				bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("嗯嗯，呵呵，我要去洗澡了")).Do();
			case *linebot.ImageMessage :
				log.Print("ImageMessage: Type(" + message.Type + "), OriginalContentURL(" + message.OriginalContentURL + "), PreviewImageURL(" + message.PreviewImageURL + ")" )
				bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("傳這甚麼廢圖？你是長輩嗎？")).Do();
			case *linebot.VideoMessage :
				log.Print("VideoMessage: Type(" + message.Type + "), OriginalContentURL(" + message.OriginalContentURL + "), PreviewImageURL(" + message.PreviewImageURL + ")" )
				bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("看甚麼影片，不知道流量快用光了嗎？")).Do();
			case *linebot.AudioMessage :
				log.Print("AudioMessage: Type(" + message.Type + "), OriginalContentURL(" + message.OriginalContentURL + "), Duration(" + message.Duration + ")" )
				bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("說的比唱的好聽，唱得鬼哭神號，是要嚇唬誰？")).Do();
			case *linebot.LocationMessage:
				log.Print("LocationMessage: Type(" + message.Type + "), Title (" + message.Title  + "), Address(" + message.Address + "), Latitude(" + message.Latitude + ")", Longitude(" + message.Longitude + ")" )
				bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("這是哪裡啊？火星嗎？")).Do();
			case *linebot.StickerMessage :
				log.Print("StickerMessage: Type(" + message.Type + "), PackageID(" + message.PackageID + "), StickerID(" + message.StickerID + ")" )
				bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("腳踏實地打字好嗎？傳這甚麼貼圖？")).Do();
			}
		}
	}
}
