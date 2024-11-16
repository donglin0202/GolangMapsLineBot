package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

		if len(data.Features) > 10 {
			data.Features = data.Features[:10]
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
		URL := "https://data.ntpc.gov.tw/api/datasets/96b6101b-c033-4834-8bd5-e312651db7a0/json?page=1&size=10"
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
		if len(records) > 10 {
			records = records[:10]
		}

		for _, item := range records {
			reply += fmt.Sprintf("案件編號: %s\n時間: %s 至 %s\n地點: %s\n\n",
				item.CaseID, item.Start, item.Stop, item.SLocation)
		}
	case "新竹市":
		URL := "https://opendata.hccg.gov.tw/API/v3/Rest/OpenData/A5C2E3B25BE2E9C2?take=10&skip=0"
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
		URL := "https://pu.hsinchu.gov.tw/svc/svc/CaseList.aspx"
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
				case 1:
					construction.Location = text
				case 4:
					construction.Period = text
				}
			})

			constructions = append(constructions, construction)
		})

		if len(constructions) > 10 {
			constructions = constructions[:10]
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

		if len(constructions) > 10 {
			constructions = constructions[:10]
		}

		for _, c := range constructions {
			reply += fmt.Sprintf("案件編號: %s\n", c.CaseNo)
			reply += fmt.Sprintf("施工地點: %s\n", c.Location)
			reply += fmt.Sprintf("工程名稱: %s\n", c.ProjectName)
			reply += fmt.Sprintf("施工期間: %s\n\n", c.Period)
		}
	default:
		reply = "目前尚未支援此縣市\n"
	}
	return reply
}
