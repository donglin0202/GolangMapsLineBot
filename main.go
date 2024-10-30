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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
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
	cb, err := webhook.ParseRequest(os.Getenv("ChannelSecret"), r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	InstructionErrorMsg := "指令格式錯誤，請重新輸入指令，支援指令格式為:\n\n"
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
	instruction := ""

	switch function {
	case "指令":
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(instruction)).Do(); err != nil {
			log.Print(err)
		}
	case "即時路況":
		if len(lines) < 3 {
			if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("指令格式錯誤，請重新輸入指令，支援指令格式為:\n\n"+instruction+"\n")).Do(); err != nil {
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
		bestRoute := getBestRoute(origin, destination)
		reply := fmt.Sprintf("從 %s 到 %s 的最佳路徑為: %s", origin, destination, bestRoute)
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
			log.Print(err)
		}

	case "預測高峰時段":
		origin := strings.TrimSpace(lines[1])
		destination := strings.TrimSpace(lines[2])
		peakTimes := getPredictedTraffic(origin, destination)
		reply := fmt.Sprintf("從 %s 到 %s 的預測高峰時段為: %s", origin, destination, peakTimes)
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
			log.Print(err)
		}

	default:
		if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage("未知的功能，請輸入: 即時路況, 最佳路徑, 或 預測高峰時段")).Do(); err != nil {
			log.Print(err)
		}
	}
}

type DirectionsResponse struct {
	Routes []struct {
		Legs []struct {
			Duration struct {
				Text string `json:"text"`
			} `json:"duration"`
			DurationInTraffic struct {
				Text string `json:"text"`
			} `json:"duration_in_traffic"`
		} `json:"legs"`
	} `json:"routes"`
}

func getTrafficCondition(origin, destination string) string {
	baseURL := "https://maps.googleapis.com/maps/api/directions/json?"
	params := url.Values{}
	params.Add("origin", origin)
	params.Add("destination", destination)
	params.Add("departure_time", "now")                 // 即時出發時間
	params.Add("language", "zh-TW")                     // 語言設定為繁體中文
	params.Add("key", os.Getenv("GOOGLE_MAPS_API_KEY")) // API key
	params.Add("traffic_model", "best_guess")           // 使用最佳交通預測模型
	params.Add("mode", "driving")                       // 交通模式為開車，走路應該不用壅塞?

	// 發送請求
	apiURL := baseURL + params.Encode()
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to send request to Google Maps API: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// 解析 API 回應
	var directionsResponse DirectionsResponse
	if err := json.Unmarshal(body, &directionsResponse); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 檢查是否有結果
	if len(directionsResponse.Routes) == 0 || len(directionsResponse.Routes[0].Legs) == 0 {
		return "無法獲取交通資訊，請確認起點和終點是否正確。"
	}

	// 取得交通狀況下的行車時間
	leg := directionsResponse.Routes[0].Legs[0]
	regularDuration := leg.Duration.Text
	trafficDuration := leg.DurationInTraffic.Text
	if regularDuration != trafficDuration {
		return fmt.Sprintf("此路段有些微壅塞\n平常開車時間:%s\n現在開車時間:%s", regularDuration, trafficDuration)
	}
	return fmt.Sprintf("交通狀況正常\n開車時間約為:%s", regularDuration)
}

func getBestRoute(origin, destination string) string {
	// 呼叫 Google Maps Directions API (最佳路徑)
	// 這裡需要替換成真實的 API 請求程式碼
	return "最佳路徑"
}

func getPredictedTraffic(origin, destination string) string {
	// 呼叫 Google Maps Traffic API (預測流量)
	// 這裡需要替換成真實的 API 請求程式碼
	return "高峰時段"
}
