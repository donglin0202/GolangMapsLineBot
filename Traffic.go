package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// 解析 API 回應
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
	if regularDuration < trafficDuration {
		return fmt.Sprintf("此路段有些微壅塞\n平常開車時間:%s\n現在開車時間:%s", regularDuration, trafficDuration)
	}
	return fmt.Sprintf("交通狀況正常\n開車時間約為:%s", regularDuration)
}
func createErrorFlexMessage(message string) map[string]interface{} {
	return map[string]interface{}{
		"type": "bubble",
		"body": map[string]interface{}{
			"type":   "box",
			"layout": "vertical",
			"contents": []map[string]interface{}{
				{
					"type":   "text",
					"text":   "發生錯誤",
					"size":   "xl",
					"color":  "#FF0000",
					"weight": "bold",
				},
				{
					"type":   "text",
					"text":   message,
					"size":   "sm",
					"color":  "#555555",
					"wrap":   true,
					"margin": "md",
				},
			},
		},
	}
}
func getBestRoute(origin, destination, mode string) map[string]interface{} {
	baseURL := "https://maps.googleapis.com/maps/api/directions/json?"
	params := url.Values{}
	params.Add("origin", origin)
	params.Add("destination", destination)
	params.Add("departure_time", "now") // 即時出發時間
	params.Add("mode", mode)
	params.Add("language", "zh-TW")           // 語言設定為繁體中文
	params.Add("traffic_model", "best_guess") // 使用最佳交通預測模型
	params.Add("key", os.Getenv("GOOGLE_MAPS_API_KEY"))

	// 發送請求
	apiURL := baseURL + params.Encode()
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Failed to send request to Google Maps API: %v", err)
		return createErrorFlexMessage("無法獲取最佳路徑")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return createErrorFlexMessage("無法讀取回應內容")
	}

	// 解析 API 回應
	type DirectionsResponse struct {
		Routes []struct {
			Legs []struct {
				Steps []struct {
					HtmlInstructions string `json:"html_instructions"`
					Duration         struct {
						Text string `json:"text"`
					} `json:"duration"`
					Distance struct {
						Text string `json:"text"`
					} `json:"distance"`
				} `json:"steps"`
			} `json:"legs"`
		} `json:"routes"`
	}
	var directionsResponse DirectionsResponse
	if err := json.Unmarshal(body, &directionsResponse); err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return createErrorFlexMessage("解析 Google Maps API 回應失敗")
	}

	// 檢查是否有路徑資料
	if len(directionsResponse.Routes) == 0 || len(directionsResponse.Routes[0].Legs) == 0 {
		return createErrorFlexMessage("無法獲取路徑資訊")
	}

	// 提取步驟資訊
	steps := directionsResponse.Routes[0].Legs[0].Steps
	removeHTMLTags := func(input string) string {
		re := regexp.MustCompile(`<[^>]*>`)   // 正則表達式匹配 HTML 標籤
		return re.ReplaceAllString(input, "") // 替換標籤為空字串
	}
	/*for idx, step := range steps {
		routeInstructions += fmt.Sprintf("\n%d. %s (需時: %s, 距離: %s)\n", idx+1, removeHTMLTags(html.UnescapeString(step.HtmlInstructions)), step.Duration.Text, step.Distance.Text)
	}

	return routeInstructions*/
	var routeSteps []map[string]interface{}

	// 將步驟整合成 Flex Message body 的內容
	for idx, step := range steps {
		routeSteps = append(routeSteps, map[string]interface{}{
			"type":   "box",
			"layout": "horizontal",
			"contents": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("%d. %s", idx+1, removeHTMLTags(html.UnescapeString(step.HtmlInstructions))),
					"size": "sm",
					"wrap": true,
				},
				{
					"type":  "text",
					"text":  fmt.Sprintf("需時: %s, 距離: %s", step.Duration.Text, step.Distance.Text),
					"size":  "xs",
					"color": "#888888",
					"align": "end",
				},
			},
		})
	}

	// 組裝完整的 Flex Message
	return map[string]interface{}{
		"type": "bubble",
		"header": map[string]interface{}{
			"type":   "box",
			"layout": "vertical",
			"contents": []map[string]interface{}{
				{
					"type":  "text",
					"text":  "FROM",
					"size":  "sm",
					"color": "#ffffff66",
				},
				{
					"type":   "text",
					"text":   origin,
					"size":   "xl",
					"color":  "#ffffff",
					"weight": "bold",
				},
				{
					"type":  "text",
					"text":  "TO",
					"size":  "sm",
					"color": "#ffffff66",
				},
				{
					"type":   "text",
					"text":   destination,
					"size":   "xl",
					"color":  "#ffffff",
					"weight": "bold",
				},
			},
			"backgroundColor": "#0367D3",
			"paddingAll":      "20px",
		},
		"body": map[string]interface{}{
			"type":     "box",
			"layout":   "vertical",
			"contents": routeSteps,
		},
	}
}

func getPredictedTraffic(origin, destination string) string {
	baseURL := "https://maps.googleapis.com/maps/api/directions/json?"
	params := url.Values{}
	params.Add("origin", origin)
	params.Add("destination", destination)
	params.Add("language", "zh-TW")
	params.Add("key", os.Getenv("GOOGLE_MAPS_API_KEY"))
	params.Add("mode", "driving")
	params.Add("traffic_model", "best_guess") // 使用最佳交通預測模型

	peakTimes := make(map[string]int)
	for hour := 0; hour < 24; hour += 2 {
		currentTime := time.Now()
		departureTime := currentTime.Add(time.Duration(hour) * time.Hour).Unix()
		params.Set("departure_time", fmt.Sprintf("%d", departureTime))

		// 發送請求
		apiURL := baseURL + params.Encode()
		resp, err := http.Get(apiURL)
		if err != nil {
			log.Printf("Failed to send request to Google Maps API: %v", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body: %v", err)
			continue
		}

		// 解析 API 回應
		type DirectionsResponse struct {
			Routes []struct {
				Legs []struct {
					DurationInTraffic struct {
						Value int `json:"value"`
					} `json:"duration_in_traffic"`
				} `json:"legs"`
			} `json:"routes"`
		}
		var directionsResponse DirectionsResponse
		if err := json.Unmarshal(body, &directionsResponse); err != nil {
			log.Printf("Failed to unmarshal response: %v", err)
			log.Printf("Response body: %s", body) // 打印回應內容以便調試
			continue
		}

		// 檢查是否有結果
		if len(directionsResponse.Routes) == 0 || len(directionsResponse.Routes[0].Legs) == 0 {
			log.Printf("No routes found in response: %s", body)
			continue
		}

		// 取得交通狀況下的行車時間
		leg := directionsResponse.Routes[0].Legs[0]
		trafficDuration := leg.DurationInTraffic.Value
		peakTimes[fmt.Sprintf("%02d:00", hour)] = trafficDuration
	}

	// 找出最壅塞的時間段
	var maxTrafficTime string
	var maxTrafficDuration int
	for time, duration := range peakTimes {
		if duration > maxTrafficDuration {
			maxTrafficDuration = duration
			maxTrafficTime = time
		}
	}

	if maxTrafficTime == "" {
		return "無法獲取預測交通資訊，請確認起點和終點是否正確。"
	}

	return fmt.Sprintf("起點: %s\n終點: %s\n預測高峰時段為: %s左右", origin, destination, maxTrafficTime)
}
