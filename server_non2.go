package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const VERSION = "v1.5.8"

var headers = http.Header{
	"clienttype":   []string{"web"},
	"content-type": []string{"application/json"},
	"user-agent":   []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"},
}

type setupG struct {
	cookie   string
	token    string
	name     string
	bank     []string
	webhookS string
	webhookF string
	spread   float64
	minCount float64
	maxCount float64
}

type globalS struct {
	count   int64
	banks   []string
	countF  int64
	orders  []string
	countSC int64
	price   float64
	delay   int64
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func getDelay(g *globalS) {
	for {
		var delay int64

		db, err := sql.Open("mysql", "root:5mmN%k48Pkl1Z3ImWXeVM9Fsl20MO$lG7GLC6awRji*2x9mP8IWDZFx}@#RX@tcp(94.103.86.233:3306)/wheel")
		if err != nil {
			panic(err.Error())
		}
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {

			}
		}(db)

		rows, err := db.Query("SELECT * FROM delay")
		if err != nil {
			panic(err.Error())
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
			}
		}(rows)

		for rows.Next() {
			err := rows.Scan(&delay)
			if err != nil {
				panic(err.Error())
			}
		}

		if err := rows.Err(); err != nil {
			panic(err.Error())
		}

		g.delay = delay
		time.Sleep(time.Hour)
	}
}

func getData(data *[]setupG, g *globalS) {
	var (
		cookie    string
		bank      string
		name      string
		spread1   string
		minCount1 string
		maxCount1 string
		token     string
		webhookS  string
		webhookF  string
		key       string
	)
	for {
		db, err := sql.Open("mysql", "root:5mmN%k48Pkl1Z3ImWXeVM9Fsl20MO$lG7GLC6awRji*2x9mP8IWDZFx}@#RX@tcp(94.103.86.233:3306)/wheel")
		if err != nil {
			panic(err.Error())
		}
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {

			}
		}(db)

		rows, err := db.Query("SELECT * FROM inf_update")
		if err != nil {
			panic(err.Error())
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
			}
		}(rows)

		for rows.Next() {
			err := rows.Scan(&key, &name, &spread1, &bank, &minCount1, &maxCount1, &token, &cookie, &webhookS, &webhookF)
			if err != nil {
				panic(err.Error())
			}

			banks := strings.Split(bank, ",")

			for _, i := range banks {
				if !contains(g.banks, i) {
					g.banks = append(g.banks, i)
				}
			}

			spread, _ := strconv.ParseFloat(spread1, 64)
			minCount, _ := strconv.ParseFloat(minCount1, 64)
			maxCount, _ := strconv.ParseFloat(maxCount1, 64)

			newItem := setupG{
				cookie:   cookie,
				token:    token,
				name:     name,
				bank:     banks,
				webhookS: webhookS,
				webhookF: webhookF,
				spread:   spread,
				minCount: minCount,
				maxCount: maxCount,
			}

			*data = append(*data, newItem)
		}

		if err := rows.Err(); err != nil {
			panic(err.Error())
		}

		time.Sleep(time.Hour)
		length := len(*data)
		for i := 0; i < length; i++ {
			(*data)[i] = setupG{}
		}
		*data = (*data)[:0]

	}
}

func main() {
	fmt.Println("WheelAIO " + VERSION)
	var (
		global  = globalS{price: 0}
		dataALL []setupG
	)

	go statusServer(&global)
	go getData(&dataALL, &global)
	go getDelay(&global)

	time.Sleep(time.Second * 5)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 5,
	}

	for {
		go tasks(client, &global, &dataALL)
		time.Sleep(time.Duration(global.delay) * time.Millisecond)
		if global.countSC != 0 {
			time.Sleep(time.Minute * 5)
		}
	}
}

func checkBanks(server []string, client []string) int {
	for _, i := range server {
		if contains(client, i) {
			return 1
		}
	}
	return 0
}

func task(data setupG, minCount float64, maxCount float64, price1 float64, price string, price1S float64, maxCountError1 float64, advNo string, banksAll []string) {
	maxCash := math.Min(maxCount, math.Min(maxCountError1, data.maxCount))

	if maxCount < data.minCount || data.maxCount < minCount || minCount > maxCount || maxCash < data.minCount || price1S-price1 < data.spread || checkBanks(banksAll, data.bank) == 0 {
		return
	}

	createOrderWebhook(data, maxCash, advNo, price, price1S, banksAll)
}

func (d *setupG) getDataSlice() []setupG {
	return []setupG{*d}
}

func tasks(client *http.Client, global *globalS, setupG *[]setupG) {
	message := map[string]interface{}{
		"asset":         "USDT",
		"fiat":          "RUB",
		"page":          1,
		"rows":          10,
		"payTypes":      global.banks,
		"merchantCheck": "False",
		"countries":     [...]string{},
		"publisherType": nil,
		"tradeType":     "BUY",
	}

	bytesRepresentation, _ := json.Marshal(message)

	resp, err := client.Post("https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Println("PROXY DIED BINANCE", err)
		global.countSC += 1
		return
	}

	var result map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println("PROXY DIED BINANCE", err)
		global.countSC += 1
		return
	}

	if resp.StatusCode != 200 {
		global.countSC += 1
	}

	data, _ := result["data"].([]interface{})

	adv, _ := data[0].(map[string]interface{})["adv"].(map[string]interface{})

	var banksAll []string
	tradeMethods, _ := adv["tradeMethods"].([]interface{})
	for _, identifier := range tradeMethods {
		banksAll = append(banksAll, identifier.(map[string]interface{})["identifier"].(string))
	}

	var minCount float64
	_, err = fmt.Sscanf(adv["minSingleTransAmount"].(string), "%f", &minCount)

	var maxCount float64
	_, err = fmt.Sscanf(adv["maxSingleTransAmount"].(string), "%f", &maxCount)

	var price1 float64
	_, err = fmt.Sscanf(adv["price"].(string), "%f", &price1)

	var maxCountError1 float64
	_, err = fmt.Sscanf(adv["dynamicMaxSingleTransAmount"].(string), "%f", &maxCountError1)

	price := adv["price"].(string)
	advNo := adv["advNo"].(string)

	for _, data1 := range *setupG {
		go task(data1, minCount, maxCount, price1, price, global.price, maxCountError1, advNo, banksAll)
	}

	advS, _ := data[1].(map[string]interface{})["adv"].(map[string]interface{})
	_, err = fmt.Sscanf(advS["price"].(string), "%f", &global.price)

	if global.count%100 == 0 {
		if global.count%10000 == 0 {
			global.count = 0
		}
		log.Printf("[SPREAD %v] [BINANCE|1 %v] [BINANCE|0 %v]", math.Floor((global.price-price1)*100)/100, global.price, price1)
	}

	global.countSC = 0
	global.count += 1
}

func createOrderWebhook(data setupG, maxCash float64, advNo string, price string, price1S float64, banksAll []string) {
	message := map[string]interface{}{
		"advOrderNumber": advNo,
		"asset":          "USDT",
		"buyType":        "BY_MONEY",
		"fiatUnit":       "RUB",
		"matchPrice":     price,
		"origin":         "MAKE_TAKE",
		"totalAmount":    maxCash,
		"tradeType":      "BUY",
	}

	bytesRepresentation, _ := json.Marshal(message)

	req, err := http.NewRequest("POST", "https://p2p.binance.com/bapi/c2c/v2/private/c2c/order-match/makeOrder", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Println(err)
	}

	req.Header = headers.Clone()
	req.Header.Set("cookie", data.cookie)
	req.Header.Set("csrftoken", data.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}

	var jsonMap map[string]interface{}
	if err = json.Unmarshal(body, &jsonMap); err != nil {
		log.Println(err)
	}

	if jsonMap["success"].(bool) == true {
		log.Println("SUCCESS")
		successWebhook(data, message, jsonMap, price1S, price, banksAll)
	} else {
		log.Println("FAIL")
		failedWebhook(data, message, jsonMap, price1S, price, banksAll)
	}
}

func successWebhook(data setupG, data1 map[string]interface{}, data2 map[string]interface{}, price1S float64, price string, banksAll []string) {
	orderNumber := data2["data"].(map[string]interface{})["orderMatch"].(map[string]interface{})["orderNumber"].(string)
	formattedTime := data2["data"].(map[string]interface{})["orderMatch"].(map[string]interface{})["createTime"].(float64)
	timeCreate := strconv.FormatFloat(formattedTime, 'f', -1, 64)

	formattedTime = data2["data"].(map[string]interface{})["orderMatch"].(map[string]interface{})["notifyPayEndTime"].(float64)
	timeEnd := strconv.FormatFloat(formattedTime, 'f', -1, 64)[:10]

	link := "https://p2p.binance.com/ru/fiatOrderDetail?orderNo=" + orderNumber + "&createdAt=" + timeCreate

	totalAmount := data1["totalAmount"].(float64)

	amount := strconv.FormatFloat(totalAmount, 'f', -1, 64)

	price1, _ := strconv.ParseFloat(price, 64)

	profit1 := totalAmount * (price1S - price1) / price1
	profit := strconv.FormatFloat(profit1, 'f', -1, 64)
	if len(profit) > 9 {
		profit = profit[:9]
	}

	spread1 := price1S - price1
	spread := strconv.FormatFloat(spread1, 'f', -1, 64)

	if len(spread) > 7 {
		spread = spread[:7]
	}

	paymentMethods := strings.Join(banksAll, ", ")

	data3 := map[string]interface{}{
		"username":   "WheelAIO",
		"avatar_url": "https://cdn.discordapp.com/attachments/1008015246134353992/1089500032023138314/Nebula.png",
		"content":    "",
		"embeds": []map[string]interface{}{
			{
				"title":     "Successful creation order Binance",
				"url":       link,
				"color":     "3800852",
				"author":    map[string]interface{}{},
				"image":     map[string]interface{}{},
				"thumbnail": map[string]interface{}{},
				"footer": map[string]string{
					"text": VERSION,
				},
				"fields": []map[string]interface{}{
					{
						"name":  "Name account",
						"value": data.name,
					},
					{
						"name":  "Amount",
						"value": amount,
					},
					{
						"name":  "Potential profit",
						"value": profit,
					},
					{
						"name":  "Spread",
						"value": spread,
					},
					{
						"name":  "Price",
						"value": price,
					},
					{
						"name":  "Valid until",
						"value": "<t:" + timeEnd + ">",
					},
					{
						"name":  "Crypto-Fiat",
						"value": data1["asset"].(string) + "-" + data1["fiatUnit"].(string),
					},
					{
						"name":  "Payment methods",
						"value": paymentMethods,
					},
				},
			},
		},
		"components": []map[string]interface{}{},
	}

	bytesRepresentation, err := json.Marshal(data3)

	if err != nil {
		log.Println(err)
		return
	}

	_, err = http.Post(data.webhookS, "application/json", bytes.NewBuffer(bytesRepresentation))

	time.Sleep(3 * time.Second)

	if err != nil {
		log.Println(err)
		return
	}

	data3 = map[string]interface{}{
		"username":   "WheelAIO",
		"avatar_url": "https://cdn.discordapp.com/attachments/1008015246134353992/1089500032023138314/Nebula.png",
		"content":    "",
		"embeds": []map[string]interface{}{
			{
				"title":     "Successful creation order Binance",
				"url":       link,
				"color":     "3800852",
				"author":    map[string]interface{}{},
				"image":     map[string]interface{}{},
				"thumbnail": map[string]interface{}{},
				"footer": map[string]string{
					"text": VERSION,
				},
				"fields": []map[string]interface{}{
					{
						"name":  "Amount",
						"value": amount,
					},
					{
						"name":  "Potential profit",
						"value": profit,
					},
					{
						"name":  "Spread",
						"value": spread,
					},
					{
						"name":  "Price",
						"value": price,
					},
					{
						"name":  "Crypto-Fiat",
						"value": data1["asset"].(string) + "-" + data1["fiatUnit"].(string),
					},
					{
						"name":  "Payment methods",
						"value": paymentMethods,
					},
				},
			},
		},
		"components": []map[string]interface{}{},
	}

	bytesRepresentation, err = json.Marshal(data3)

	if err != nil {
		log.Println(err)
		return
	}

	_, err = http.Post("https://discord.com/api/webhooks/999004301462609981/oHYKy2_Z30-BLqc8eU3tUb9kI8uo6s9EAmIK3bt1BZh2qZ492GUTRi3lc27kzuL5CS7P", "application/json", bytes.NewBuffer(bytesRepresentation))

	if err != nil {
		log.Println(err)
		return
	}
}

func failedWebhook(data setupG, data1 map[string]interface{}, data2 map[string]interface{}, price1S float64, price string, banksAll []string) {
	time.Sleep(3 * time.Second)

	totalAmount := data1["totalAmount"].(float64)

	amount := strconv.FormatFloat(totalAmount, 'f', -1, 64)

	price1, _ := strconv.ParseFloat(price, 64)

	profit1 := totalAmount * (price1S - price1) / price1
	profit := strconv.FormatFloat(profit1, 'f', -1, 64)

	if len(profit) > 9 {
		profit = profit[:9]
	}

	spread1 := price1S - price1
	spread := strconv.FormatFloat(spread1, 'f', -1, 64)

	if len(spread) > 7 {
		spread = spread[:7]
	}

	paymentMethods := strings.Join(banksAll, ", ")

	data3 := map[string]interface{}{
		"username":   "WheelAIO",
		"avatar_url": "https://cdn.discordapp.com/attachments/1008015246134353992/1089500032023138314/Nebula.png",
		"content":    "",
		"embeds": []map[string]interface{}{
			{
				"title":     "Failed creation order Binance",
				"color":     "16252928",
				"author":    map[string]interface{}{},
				"image":     map[string]interface{}{},
				"thumbnail": map[string]interface{}{},
				"footer": map[string]string{
					"text": VERSION,
				},
				"fields": []map[string]interface{}{
					{
						"name":  "Error",
						"value": data2["message"].(string),
					},
					{
						"name":  "Name account",
						"value": data.name,
					},
					{
						"name":  "Amount",
						"value": amount,
					},
					{
						"name":  "Potential profit",
						"value": profit,
					},
					{
						"name":  "Spread",
						"value": spread,
					},
					{
						"name":  "Price",
						"value": price,
					},
					{
						"name":  "Crypto-Fiat",
						"value": data1["asset"].(string) + "-" + data1["fiatUnit"].(string),
					},
					{
						"name":  "Payment methods",
						"value": paymentMethods,
					},
				},
			},
		},
		"components": []map[string]interface{}{},
	}

	bytesRepresentation, err := json.Marshal(data3)

	if err != nil {
		log.Println(err)
		return
	}

	_, err = http.Post(data.webhookF, "application/json", bytes.NewBuffer(bytesRepresentation))
}

func statusServer(global *globalS) {
	ip := getIp()

	var (
		active     string
		lastUpdate string
		ip1        string
	)

	for {
		checkIp := 0
		db, err := sql.Open("mysql", "root:5mmN%k48Pkl1Z3ImWXeVM9Fsl20MO$lG7GLC6awRji*2x9mP8IWDZFx}@#RX@tcp(94.103.86.233:3306)/wheel")
		if err != nil {
			panic(err.Error())
		}
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {

			}
		}(db)

		rows, err := db.Query("SELECT * FROM inf_servers")
		if err != nil {
			panic(err.Error())
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {

			}
		}(rows)

		for rows.Next() {
			err := rows.Scan(&ip1, &active, &lastUpdate)
			if err != nil {
				panic(err.Error())
			}

			if ip1 == ip {
				checkIp = 1
			}
		}

		if checkIp == 1 {
			query := "UPDATE inf_servers SET active =?, last_update =? WHERE ip = ?"
			_, err = db.Exec(query, global.countSC, time.Now().Format("2006-01-02 15:04:05"), ip)
			if err != nil {
				log.Fatal(err)
			}

		} else {
			query := "INSERT inf_servers (ip, active, last_update) VALUES (?, ?, ?)"
			_, err = db.Exec(query, ip, global.countSC, time.Now().Format("2006-01-02 15:04:05"))
		}

		if err := rows.Err(); err != nil {
			panic(err.Error())
		}

		time.Sleep(time.Hour)
	}
}

func getIp() string {
	resp, _ := http.Get("https://api.ipify.org")
	ip, _ := ioutil.ReadAll(resp.Body)

	if len(ip) > 14 {
		ip = ip[:14]
	}

	return string(ip)
}
