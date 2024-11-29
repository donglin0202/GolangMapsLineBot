package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GetConstruction(target string) string {
	var reply string
	target = strings.Replace(target, "臺", "台", -1)
	switch target {
	case "台北市":
		URL := "https://tpnco.blob.core.windows.net/blobfs/Todaywork.json"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("無法取得資料:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取資料錯誤:", err)
			return "Error，請再試一次"
		}

		var data struct {
			Features []struct {
				Properties struct {
					Ac_no   string `json:"Ac_no"`
					AppTime string `json:"AppTime"`
					Addr    string `json:"Addr"`
					AppMode string `json:"AppMode"`
				} `json:"properties"`
			} `json:"features"`
		}
		body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf")) // 移除 BOM
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("解析 JSON 錯誤:", err)
			return "Error，請再試一次"
		}

		modeMap := map[string]string{
			"0": "施工通報",
			"3": "銑鋪通報",
			"4": "搶修通報",
			"5": "道路維護通報",
			"6": "人手孔施工通報",
			"B": "建案公設復舊",
		}

		if len(data.Features) > 5 {
			data.Features = data.Features[:5]
		}

		for _, feature := range data.Features {
			p := feature.Properties
			modeName, ok := modeMap[p.AppMode]
			if !ok {
				modeName = "未知類別"
			}
			reply += fmt.Sprintf("案件編號: %s\n時間: %s\n地點: %s\n通報類別: %s\n\n", p.Ac_no, p.AppTime, p.Addr, modeName)
		}
	case "新北市":
		URL := "https://data.ntpc.gov.tw/api/datasets/96b6101b-c033-4834-8bd5-e312651db7a0/json?page=1&size=5"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("無法取得資料:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取資料錯誤:", err)
			return "Error，請再試一次"
		}

		var data []struct {
			CaseID       string `json:"CaseID"`
			CaseName     string `json:"案件名稱"`
			DigSite      string `json:"DigSite"`
			CaseStart    string `json:"CaseStart"`
			CaseEnd      string `json:"CaseEnd"`
			StrAllowTime string `json:"StrAllowTime"`
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("解析 JSON 錯誤:", err)
			return "Error，請再試一次"
		}

		for _, item := range data {
			reply += fmt.Sprintf("案件名稱: %s\n時間: %s 至 %s\n地點: %s\n\n",
				item.CaseName, item.CaseStart, item.CaseEnd, item.DigSite)
		}
	case "桃園市":
		URL := "http://data.tycg.gov.tw/api/v1/rest/datastore/52de3762-1490-4a86-a074-0062d746873b?format=json"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("無法取得資料:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取資料錯誤:", err)
			return "Error，請再試一次"
		}

		var result struct {
			Success bool `json:"success"`
			Result  struct {
				Records []struct {
					CaseID    string `json:"CaseID"`
					Start     string `json:"Start"`
					Stop      string `json:"stop"`
					SLocation string `json:"SLocation"`
				} `json:"records"`
			} `json:"result"`
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Println("解析 JSON 錯誤:", err)
			return "Error，請再試一次"
		}

		records := result.Result.Records
		if len(records) > 5 {
			records = records[:5]
		}

		for _, item := range records {
			reply += fmt.Sprintf("案件編號: %s\n時間: %s 至 %s\n地點: %s\n\n",
				item.CaseID, item.Start, item.Stop, item.SLocation)
		}
	case "新竹市":
		URL := os.Getenv("PROXY_URL") + "https://opendata.hccg.gov.tw/API/v3/Rest/OpenData/A5C2E3B25BE2E9C2?take=5&skip=0"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("無法取得資料:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取資料錯誤:", err)
			return "Error，請再試一次"
		}

		var data []struct {
			UnitName    string `json:"單位名稱"`
			Percentage  string `json:"百分比"`
			TotalLength string `json:"總長度"`
			TotalArea   string `json:"總面積"`
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("解析 JSON 錯誤:", err)
			return "Error，請再試一次"
		}

		for _, item := range data {
			reply += fmt.Sprintf("名稱: %s\n百分比: %s\n", item.UnitName, item.Percentage)

			if item.TotalLength == "0" {
				reply += "總長度太小無法統計\n"
			} else {
				reply += fmt.Sprintf("總長度: %s\n", item.TotalLength)
			}

			if item.TotalArea == "0" {
				reply += "總面積太小無法統計\n\n"
			} else {
				reply += fmt.Sprintf("總面積: %s\n\n", item.TotalArea)
			}
		}
	case "新竹縣":
		URL := os.Getenv("PROXY_URL") + "https://pu.hsinchu.gov.tw/svc/svc/CaseList.aspx"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		// 使用 goquery 解析 HTML
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println("解析 HTML 失敗:", err)
			return "Error，請再試一次"
		}

		type Construction struct {
			CaseNo   string // 案件編號
			Location string // 施工位置
			Period   string // 施工期間
		}

		constructions := []Construction{}

		doc.Find("table.GridViewCss tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}

			var construction Construction
			s.Find("td").Each(func(j int, td *goquery.Selection) {
				text := strings.TrimSpace(td.Text())
				switch j {
				case 0:
					construction.CaseNo = text
				case 3:
					construction.Location = text
				case 4:
					construction.Period = text
				}
			})

			constructions = append(constructions, construction)
		})

		if len(constructions) > 5 {
			constructions = constructions[:5]
		}

		for _, c := range constructions {
			reply += fmt.Sprintf("案件編號: %s\n", c.CaseNo)
			reply += fmt.Sprintf("施工期間: %s\n", c.Period)
			reply += fmt.Sprintf("施工位置: %s\n\n", c.Location)
		}
	case "苗栗縣":
		URL := "https://miaoli-road.miaoli.gov.tw/NewMiaoliWeb/Common/CaseList_New.aspx"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println("解析 HTML 失敗:", err)
			return "Error，請再試一次"
		}

		type Construction struct {
			CaseNo      string // 案件編號
			Location    string // 施工地點
			ProjectName string // 工程名稱
			Period      string // 施工期間
		}

		constructions := []Construction{}

		doc.Find("table.gview tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}

			var construction Construction
			var startDate, endDate string

			s.Find("td").Each(func(j int, td *goquery.Selection) {
				text := strings.TrimSpace(td.Text())
				switch j {
				case 0:
					construction.Location = text
				case 1:
					construction.CaseNo = text
				case 3:
					construction.ProjectName = text
				case 4:
					startDate = text
				case 5:
					endDate = text
				}
			})

			if startDate != "" && endDate != "" {
				construction.Period = fmt.Sprintf("%s ~ %s", startDate, endDate)
			}

			constructions = append(constructions, construction)
		})

		if len(constructions) > 5 {
			constructions = constructions[:5]
		}

		for _, c := range constructions {
			reply += fmt.Sprintf("案件編號: %s\n", c.CaseNo)
			reply += fmt.Sprintf("施工地點: %s\n", c.Location)
			reply += fmt.Sprintf("工程名稱: %s\n", c.ProjectName)
			reply += fmt.Sprintf("施工期間: %s\n\n", c.Period)
		}
	case "台中市":
		URL := "https://datacenter.taichung.gov.tw/swagger/OpenData/d5adb71a-00bb-4573-b67e-ffdccfc7cd27?limit=5"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		type Construction struct {
			CaseNo      string `json:"申請書編號"`
			Location    string `json:"地點"`
			ProjectName string `json:"工程名稱"`
			StartDate   string `json:"核准起日"`
			EndDate     string `json:"核准迄日"`
		}

		var constructions []Construction
		err = json.NewDecoder(resp.Body).Decode(&constructions)
		if err != nil {
			fmt.Println("解析 JSON 失敗:", err)
			return "Error，請再試一次"
		}

		for _, c := range constructions {
			period := fmt.Sprintf("%s ~ %s", c.StartDate, c.EndDate)
			reply += fmt.Sprintf("編號: %s\n", c.CaseNo)
			reply += fmt.Sprintf("地點: %s\n", c.Location)
			reply += fmt.Sprintf("名稱: %s\n", c.ProjectName)
			reply += fmt.Sprintf("時間: %s\n\n", period)
		}
	case "嘉義市":
		URL := "https://data.chiayi.gov.tw/opendata/api/getResource?oid=2cf1aa4f-3cdd-46a0-be84-b6f161cd892d&rid=ca4c025d-1856-4200-81b7-a401aa653da3"
		client := &http.Client{}
		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			fmt.Println("建立請求失敗:", err)
			return "Error，請再試一次"
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) "+
			"Chrome/58.0.3029.110 Safari/537.3")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取回應失敗:", err)
			return "Error，請再試一次"
		}

		type CaseDetail struct {
			CaseID    string `xml:"CASE_ID"`
			ConstName string `xml:"CONST_NAME"`
			Location  string `xml:"LOCATION"`
			ABE_DA    string `xml:"ABE_DA"`
			AEN_DA    string `xml:"AEN_DA"`
		}

		type CaseList struct {
			CaseDetails []CaseDetail `xml:"CASE_DETAIL"`
		}

		type DigCase struct {
			CaseList CaseList `xml:"CASE_LIST"`
		}

		var digCase DigCase
		err = xml.Unmarshal(bodyBytes, &digCase)
		if err != nil {
			fmt.Println("解析 XML 失敗:", err)
			return "Error，請再試一次"
		}

		constructions := digCase.CaseList.CaseDetails
		if len(constructions) > 5 {
			constructions = constructions[:5]
		}

		for _, c := range constructions {
			reply += fmt.Sprintf("編號: %s\n", c.CaseID)
			reply += fmt.Sprintf("地點: %s\n", c.Location)
			reply += fmt.Sprintf("名稱: %s\n", c.ConstName)
			reply += fmt.Sprintf("時間: %s ~ %s\n\n", c.ABE_DA, c.AEN_DA)
		}
	case "嘉義縣":
		URL := os.Getenv("PROXY_URL") + "https://publicpipe.cyhg.gov.tw/ChiayiPub/Report1.aspx"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP 請求失敗，狀態碼: %d\n", resp.StatusCode)
			return "Error，請再試一次"
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println("解析 HTML 失敗:", err)
			return "Error，請再試一次"
		}

		type ChiayiCountyConstruction struct {
			StartPoint string
			EndPoint   string
			Date       string
			Status     string
		}

		constructions := []ChiayiCountyConstruction{}

		doc.Find("table#ctl00_ContentPlaceHolder1_GridView1 tr.cssRow").Each(func(i int, s *goquery.Selection) {
			if len(constructions) >= 5 {
				return
			}

			tds := s.Find("td")
			if tds.Length() < 5 {
				return
			}

			// 提取起點和終點
			startSpan := s.Find("span[id$='LabDigStart']").First()
			endSpan := s.Find("span[id$='LabDigEnd']").First()

			startText := strings.TrimSpace(startSpan.Text())
			endText := strings.TrimSpace(endSpan.Text())
			// 將第二個起點替換為終點
			endText = strings.Replace(endText, "起點：", "終點：", 1)

			// 提取日期
			dateText := strings.TrimSpace(tds.Eq(3).Text())
			startDate := ""
			endDate := ""
			if strings.Contains(dateText, "至") {
				parts := strings.SplitN(dateText, "至", 2)
				if len(parts) == 2 {
					startDate = strings.TrimSpace(strings.TrimPrefix(parts[0], "自"))
					endDate = strings.TrimSpace(parts[1])
				}
			}
			combinedDate := ""
			if startDate != "" && endDate != "" {
				combinedDate = fmt.Sprintf("%s至%s", startDate, endDate)
			} else {
				combinedDate = dateText
			}

			// 提取狀態
			status := strings.TrimSpace(tds.Eq(4).Text())

			constructions = append(constructions, ChiayiCountyConstruction{
				StartPoint: startText,
				EndPoint:   endText,
				Date:       combinedDate,
				Status:     status,
			})
		})

		for _, c := range constructions {
			reply += fmt.Sprintf("%s\n", c.StartPoint)
			reply += fmt.Sprintf("%s\n", c.EndPoint)
			reply += fmt.Sprintf("日期: %s\n", c.Date)
			reply += fmt.Sprintf("狀態: %s\n\n", c.Status)
		}
	case "台南市":
		URL := "https://soa.tainan.gov.tw/Api/Service/Get/8d5e8855-b9c4-4acd-b73d-cf2afb2e51e5"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP 請求失敗，狀態碼: %d\n", resp.StatusCode)
			return "Error，請再試一次"
		}

		type TainanEntry struct {
			Location string `json:"LOCATION"`
			ABE_DA   string `json:"ABE_DA"`
			AEN_DA   string `json:"AEN_DA"`
			StatDesc string `json:"StatDesc"`
		}

		type TainanResponse struct {
			Data []TainanEntry `json:"data"`
		}

		var tainanData TainanResponse
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取回應失敗:", err)
			return "Error，請再試一次"
		}

		err = json.Unmarshal(bodyBytes, &tainanData)
		if err != nil {
			fmt.Println("解析 JSON 失敗:", err)
			return "Error，請再試一次"
		}

		conns := tainanData.Data

		// 將資料按照 ABE_DA (開始日期) 進行排序，最新的在前
		sort.Slice(conns, func(i, j int) bool {
			// 提取並分割日期部分
			startDateStr := strings.Split(strings.TrimSpace(conns[i].ABE_DA), " ")[0]

			timeI, err1 := time.Parse("2006/1/2", startDateStr)
			if err1 != nil {
				// 解析失敗
				return false
			}

			startDateStrJ := strings.Split(strings.TrimSpace(conns[j].ABE_DA), " ")[0]

			timeJ, err2 := time.Parse("2006/1/2", startDateStrJ)
			if err2 != nil {
				return false
			}

			return timeI.After(timeJ)
		})

		// 取最新的10筆資料
		if len(conns) > 5 {
			conns = conns[:5]
		}

		for _, item := range conns {
			// 分割起點與終點
			locations := strings.Split(item.Location, "訖")
			startPoint := ""
			endPoint := ""
			if len(locations) == 2 {
				startPoint = strings.TrimSpace(locations[0])
				endPoint = strings.TrimSpace(locations[1])
			} else {
				startPoint = strings.TrimSpace(item.Location)
				endPoint = "無資料"
			}

			// 日期
			startDateStr := strings.Split(strings.TrimSpace(item.ABE_DA), " ")[0]
			endDateStr := strings.Split(strings.TrimSpace(item.AEN_DA), " ")[0]
			var combinedDate string

			// 將日期轉換為民國年格式
			startDate, err1 := time.Parse("2006/1/2", startDateStr)
			if err1 != nil {
				fmt.Printf("解析開始日期失敗: %v\n", err1)
				startDate = time.Time{}
			}
			endDate, err2 := time.Parse("2006/1/2", endDateStr)
			if err2 != nil {
				fmt.Printf("解析結束日期失敗: %v\n", err2)
				endDate = time.Time{}
			}

			if !startDate.IsZero() && !endDate.IsZero() {
				combinedDate = fmt.Sprintf("%d年%02d月%02d日至%d年%02d月%02d日",
					startDate.Year()-1911, startDate.Month(), startDate.Day(),
					endDate.Year()-1911, endDate.Month(), endDate.Day())
			} else if !startDate.IsZero() {
				combinedDate = fmt.Sprintf("%d年%02d月%02d日至",
					startDate.Year()-1911, startDate.Month(), startDate.Day())
			} else {
				combinedDate = "無資料"
			}

			// 狀態
			status := strings.TrimSpace(item.StatDesc)
			if status == "" {
				status = "無狀態資料"
			}

			reply += fmt.Sprintf("起點：%s\n", startPoint)
			reply += fmt.Sprintf("終點：%s\n", endPoint)
			reply += fmt.Sprintf("日期: %s\n", combinedDate)
			reply += fmt.Sprintf("狀態: %s\n\n", status)
		}
	case "高雄市":
		URL := os.Getenv("PROXY_URL") + "https://pipegis.kcg.gov.tw/openDataService.aspx"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP 請求失敗，狀態碼: %d\n", resp.StatusCode)
			return "Error，請再試一次"
		}

		type DigCaseInfo struct {
			Location string `xml:"LOCATION"`
			DateDigS string `xml:"DATE_DIG_S"`
			DateDigE string `xml:"DATE_DIG_E"`
			Reason   string `xml:"REASON"`
		}

		type DigCaseInfos struct {
			XMLName     xml.Name      `xml:"DigCaseInfos"`
			DigCaseInfo []DigCaseInfo `xml:"DigCaseInfo"`
		}

		var kaohsiungData DigCaseInfos
		err = xml.NewDecoder(resp.Body).Decode(&kaohsiungData)
		if err != nil {
			fmt.Println("解析 XML 失敗:", err)
			return "Error，請再試一次"
		}

		conns := kaohsiungData.DigCaseInfo

		if len(conns) > 5 {
			conns = conns[:5]
		}

		for _, item := range conns {
			address := strings.TrimSpace(item.Location)
			startDateStr := strings.TrimSpace(item.DateDigS)
			endDateStr := strings.TrimSpace(item.DateDigE)
			var combinedDate string

			if len(startDateStr) == 7 && len(endDateStr) == 7 {
				rocStartYear, err1 := strconv.Atoi(startDateStr[:3])
				startMonth, err2 := strconv.Atoi(startDateStr[3:5])
				startDay, err3 := strconv.Atoi(startDateStr[5:7])
				rocEndYear, err4 := strconv.Atoi(endDateStr[:3])
				endMonth, err5 := strconv.Atoi(endDateStr[3:5])
				endDay, err6 := strconv.Atoi(endDateStr[5:7])

				if err1 == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil && err6 == nil {
					combinedDate = fmt.Sprintf("%d年%02d月%02d日至%d年%02d月%02d日",
						rocStartYear, startMonth, startDay,
						rocEndYear, endMonth, endDay)
				} else {
					combinedDate = "無資料"
				}
			} else {
				combinedDate = "無資料"
			}

			description := strings.TrimSpace(item.Reason)
			if description == "" {
				description = "無說明資料"
			}

			reply += fmt.Sprintf("地址：%s\n", address)
			reply += fmt.Sprintf("時間: %s\n", combinedDate)
			reply += fmt.Sprintf("說明: %s\n\n", description)
		}
	case "屏東縣":
		URL := os.Getenv("PROXY_URL") + "https://cdn.odportal.tw/api/v1/resource/SfsCLK-2/61ae820f3a1469002467f7c3"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP 請求失敗，狀態碼: %d\n", resp.StatusCode)
			return "Error，請再試一次"
		}

		type PingtungCase struct {
			Applicant    string `xml:"申請單位"`
			ApprovalUnit string `xml:"核准機關"`
			Reason       string `xml:"施工原因"`
			Location     string `xml:"挖掘地點"`
			PermitNumber string `xml:"道路挖掘許可證字號"`
			StartDate    string `xml:"核准施工起始日期"`
			EndDate      string `xml:"核准施工終止日期"`
		}

		type PingtungData struct {
			XMLName xml.Name       `xml:"屏東縣道路挖掘施工資訊"`
			Cases   []PingtungCase `xml:"屏東縣道路挖掘施工案件"`
		}

		var pingtungData PingtungData
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取回應失敗:", err)
			return "Error，請再試一次"
		}

		err = xml.Unmarshal(bodyBytes, &pingtungData)
		if err != nil {
			fmt.Println("解析 XML 失敗:", err)
			return "Error，請再試一次"
		}

		cases := pingtungData.Cases

		if len(cases) > 5 {
			cases = cases[:5]
		}

		for _, item := range cases {
			startDateStr := strings.TrimSpace(item.StartDate)
			endDateStr := strings.TrimSpace(item.EndDate)
			var combinedDate string

			if len(startDateStr) == 7 && len(endDateStr) == 7 {
				rocStartYear, err1 := strconv.Atoi(startDateStr[:3])
				startMonth, err2 := strconv.Atoi(startDateStr[3:5])
				startDay, err3 := strconv.Atoi(startDateStr[5:7])
				rocEndYear, err4 := strconv.Atoi(endDateStr[:3])
				endMonth, err5 := strconv.Atoi(endDateStr[3:5])
				endDay, err6 := strconv.Atoi(endDateStr[5:7])

				if err1 == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil && err6 == nil {
					combinedDate = fmt.Sprintf("%d年%02d月%02d日至%d年%02d月%02d日",
						rocStartYear, startMonth, startDay,
						rocEndYear, endMonth, endDay)
				} else {
					combinedDate = "無資料"
				}
			} else {
				combinedDate = "無資料"
			}

			reason := strings.TrimSpace(item.Reason)
			if reason == "" {
				reason = "無原因資料"
			}

			location := strings.TrimSpace(item.Location)
			if location == "" {
				location = "無地點資料"
			}

			reply += fmt.Sprintf("原因：%s\n", reason)
			reply += fmt.Sprintf("地點：%s\n", location)
			reply += fmt.Sprintf("日期：%s\n\n", combinedDate)
		}
	case "宜蘭縣":
		URL := os.Getenv("PROXY_URL") + "https://cdn.odportal.tw/api/v1/resource/DSNTMGUM/61b504bb6e97860024674b09"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP 請求失敗，狀態碼: %d\n", resp.StatusCode)
			return "Error，請再試一次"
		}

		type YilanCase struct {
			Name      string `xml:"CONST_NAME"`
			Location  string `xml:"LOCATION"`
			StartDate string `xml:"ABE_DA"`
			EndDate   string `xml:"AEN_DA"`
		}

		type YilanData struct {
			XMLName xml.Name    `xml:"DIG_CASE"`
			Cases   []YilanCase `xml:"CASE_LIST>CASE_DETAIL"`
		}

		var yilanData YilanData
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取回應失敗:", err)
			return "Error，請再試一次"
		}

		err = xml.Unmarshal(bodyBytes, &yilanData)
		if err != nil {
			fmt.Println("解析 XML 失敗:", err)
			return "Error，請再試一次"
		}

		cases := yilanData.Cases

		if len(cases) > 5 {
			cases = cases[:5]
		}

		for _, item := range cases {
			startDateStr := strings.TrimSpace(item.StartDate)
			endDateStr := strings.TrimSpace(item.EndDate)
			var combinedDate string

			startDate, err1 := time.Parse("2006-01-02", startDateStr)
			endDate, err2 := time.Parse("2006-01-02", endDateStr)

			if err1 == nil && err2 == nil {
				combinedDate = fmt.Sprintf("%d年%02d月%02d日至%d年%02d月%02d日",
					startDate.Year()-1911, startDate.Month(), startDate.Day(),
					endDate.Year()-1911, endDate.Month(), endDate.Day())
			} else if err1 == nil {
				combinedDate = fmt.Sprintf("%d年%02d月%02d日至",
					startDate.Year()-1911, startDate.Month(), startDate.Day())
			} else {
				combinedDate = "無資料"
			}

			name := strings.TrimSpace(item.Name)
			if name == "" {
				name = "無名稱資料"
			}

			location := strings.TrimSpace(item.Location)
			if location == "" {
				location = "無地點資料"
			}

			reply += fmt.Sprintf("名稱：%s\n", name)
			reply += fmt.Sprintf("地點：%s\n", location)
			reply += fmt.Sprintf("日期：%s\n\n", combinedDate)
		}
	case "花蓮縣":
		URL := "https://pipe.hl.gov.tw/hualienpipe/Pub/PubQuery.aspx"
		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println("解析 HTML 失敗:", err)
			return "Error，請再試一次"
		}

		type HualienConstruction struct {
			Unit     string // 施工單位
			Date     string // 施工日期
			Location string // 施工地點
		}

		var constructions []HualienConstruction

		doc.Find("table[id^='ctl00_ContentPlaceHolder1_CList_ctl'][id$='Table1']").Each(func(i int, s *goquery.Selection) {
			tds := s.Find("td")
			if tds.Length() < 5 {
				return
			}

			unit := strings.TrimSpace(tds.Eq(1).Text())
			date := strings.TrimSpace(tds.Eq(2).Text())
			locationNode := tds.Eq(4)
			locationHtml, err := locationNode.Html()
			if err != nil {
				locationHtml = locationNode.Text()
			}
			locationHtml = strings.ReplaceAll(locationHtml, "<br />", "\n")
			locationHtml = strings.ReplaceAll(locationHtml, "<br/>", "\n")
			locationHtml = strings.ReplaceAll(locationHtml, "<br>", "\n")
			locationHtml = html.UnescapeString(locationHtml)

			// 移除所有 HTML 標籤
			locationText := regexp.MustCompile(`\<.*?\>`).ReplaceAllString(locationHtml, "")
			locationText = strings.TrimSpace(locationText)

			// 提取施工地點，忽略 "合約編號：" 和 "其他事項："
			locationLines := strings.Split(locationText, "\n")
			var location string
			for _, line := range locationLines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "合約編號：") && !strings.HasPrefix(line, "其他事項：") {
					location = line
					break
				}
			}

			if location == "" || location == "請選擇請選擇" {
				location = "無地點資料"
			}

			constructions = append(constructions, HualienConstruction{
				Unit:     unit,
				Date:     date,
				Location: location,
			})
		})

		if len(constructions) > 7 {
			constructions = constructions[2:7]
		} else if len(constructions) > 2 {
			constructions = constructions[2:]
		}

		for _, c := range constructions {
			reply += fmt.Sprintf("施工單位：%s\n", c.Unit)
			reply += fmt.Sprintf("施工日期：%s\n", c.Date)
			reply += fmt.Sprintf("施工地點：%s\n\n", c.Location)
		}
	case "金門縣":
		now := time.Now()
		sdate := now.AddDate(0, 0, -30).Format("2006-01-02")
		edate := now.AddDate(0, 0, 30).Format("2006-01-02")
		URL := fmt.Sprintf("https://roaddig.kinmen.gov.tw/KMDigAPI/api/OpenData/GetCaseList?sdate=%s&edate=%s", sdate, edate)

		resp, err := http.Get(URL)
		if err != nil {
			fmt.Println("HTTP GET 失敗:", err)
			return "Error，請再試一次"
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP 請求失敗，狀態碼: %d\n", resp.StatusCode)
			return "Error，請再試一次"
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("讀取回應失敗:", err)
			return "Error，請再試一次"
		}

		type PeriodCaseData struct {
			EngUse     string `json:"EngUse"`
			Road       string `json:"Road"`
			SchedStart string `json:"SchedStart"`
			SchedStop  string `json:"SchedStop"`
		}

		type OpenDataResult struct {
			IsSuccessful bool             `json:"IsSuccessful"`
			ErrorMessage *string          `json:"ErrorMessage"`
			Data         []PeriodCaseData `json:"Data"`
		}

		var kinmenData OpenDataResult
		err = json.Unmarshal(bodyBytes, &kinmenData)
		if err != nil {
			fmt.Println("解析 JSON 失敗:", err)
			return "Error，請再試一次"
		}

		if !kinmenData.IsSuccessful {
			return "Error，請再試一次"
		}

		cases := kinmenData.Data

		// 取前5筆資料
		if len(cases) > 5 {
			cases = cases[:5]
		}

		for _, item := range cases {
			engUse := strings.TrimSpace(item.EngUse)
			if engUse == "" {
				engUse = "無名稱資料"
			}

			road := strings.TrimSpace(item.Road)
			if road == "" {
				road = "無地點資料"
			}

			// 提取日期部分，去除時間
			extractDate := func(datetime string) string {
				parts := strings.Split(datetime, " ")
				if len(parts) > 0 {
					return parts[0]
				}
				return "無日期資料"
			}

			allowStart := extractDate(strings.TrimSpace(item.SchedStart))
			allowStop := extractDate(strings.TrimSpace(item.SchedStop))

			dateRange := fmt.Sprintf("%s 至 %s", allowStart, allowStop)

			reply += fmt.Sprintf("名稱：%s\n", engUse)
			reply += fmt.Sprintf("地點：%s\n", road)
			reply += fmt.Sprintf("日期：%s\n\n", dateRange)
		}
	default:
		reply = "目前尚未支援此縣市"
	}
	return reply
}
