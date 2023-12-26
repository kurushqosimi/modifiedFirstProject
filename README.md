package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type Request struct {
	XMLNAME            xml.Name `xml:"request" json:"-"`
	Command            string   `xml:"command" binding:"required" json:",omitempty"`
	Login              string   `xml:"login" json:",omitempty"`
	ServiceID          int      `xml:"serviceID" json:",omitempty"`
	Phone              string   `xml:"phone,omitempty" json:",omitempty"`
	PhoneNew           string   `xml:"phoneNew,omitempty" json:",omitempty"`
	AccountNum         string   `xml:"accountNum,omitempty" json:",omitempty"`
	Vendor             int      `xml:"vendor,omitempty" json:",omitempty"`
	RequestType        string   `xml:"requestType,omitempty" json:",omitempty"`
	Amount             float64  `xml:"amount,omitempty" json:",omitempty"`
	AmountWithCommiss  float64  `xml:"amountWithCommiss,omitempty" json:",omitempty"`
	ClientPayer        string   `xml:"clientPayer,omitempty" json:",omitempty"`
	AccountNumPayer    string   `xml:"accountNumPayer,omitempty" json:",omitempty"`
	ClientReceiver     string   `xml:"clientReceiver,omitempty" json:",omitempty"`
	AccountNumReceiver string   `xml:"accountNumReceiver,omitempty" json:",omitempty"`
	PreSharedKey       int      `xml:"preSharedKey" json:",omitempty"`
	PreCheckQueueID    int64    `xml:"precheckQueueID,omitempty" json:",omitempty"`
	TransID            string   `xml:"transID,omitempty" json:",omitempty"`
	CallID             int64    `xml:"callID,omitempty" json:",omitempty"`
	QR                 string   `xml:"qr,omitempty" json:",omitempty"`
	CardHash           string   `xml:"cardHash,omitempty" json:",omitempty"`
	ExtTransID         string   `xml:"extTransID,omitempty" json:",omitempty"`
	WalletID           int      `xml:"walletID,omitempty" json:"walletID,omitempty"`
	HashSum            string   `xml:"hashSum" json:"hashSum"`
	NotifyRoute        string   `xml:"notifyRoute" json:"notifyRoute"`
}

func GetSha512Hash(text string) string {
	sha512 := sha512.New()
	sha512.Write([]byte(text))
	return hex.EncodeToString(sha512.Sum(nil))
}

type ResponseFromXml struct {
	XMLName      xml.Name       `xml:"response"`
	Status       int64          `xml:"status"`
	Detail       string         `xml:"statusdetails"`
	PaymentID    *int64         `xml:"paymentid"`
	PrecheckInfo *Precheck      `xml:"precheckinfo"`
	ServicesInfo *[]ServiceList `xml:"servicelist>service"`
	CreatedAt    *time.Time     `xml:"created_at"`
	ProcessedAt  *time.Time     `xml:"processed_at"`
	Remain       *float64       `xml:"remain"`
	Overdraft    *float64       `xml:"overdraft"`
}

type Precheck struct {
	RawInfo string `xml"rawinfo"`
	Items   *struct {
		Name     *string `xml:"name"`
		Address  *string `xml:"address"`
		Previous *string `xml:"previous"`
		Present  *string `xml:"present"`
		Date     *string `xml:"date"`
		Rest     *string `xml:"rest"`
		Item     *[]struct {
			Label string `xml:"label,attr"`
			Value string `xml:"value,attr"`
		} `xml:"item"`
	} `xml:"items"`
}

type ServiceList struct {
	ID      int64  `xml:"id,attr"`
	Caption string `xml:"caption,attr"`
	Pattern string `xml:"pattern,attr"`
	Descr   string `xml:"descr,attr"`
}

func ValidMACByte(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return expectedMAC
}

func main() {
	sumTime = atomic.Value{}
	errCount = atomic.Value{}
	sumTime.Store(int64(0))
	errCount.Store(0)
	r := gin.New()
	r.GET("/", mainRoute)

	r.GET("/w2", simple)
	/*	port := os.Args[1]
		if port == "" {*/
	port := "0001"
	//	}
	r.Run(":" + port)

}
func simple(c *gin.Context) {
	sumTime = atomic.Value{}
	errCount = atomic.Value{}
	sumTime.Store(int64(0))
	errCount.Store(0)
	var wg sync.WaitGroup
	psID := c.Query("ps_id")
	key := c.Query("key")
	number := c.Query("number")
	types := c.Query("type")
	// amount := c.Query("amount")
	// amount2credit := c.Query("amount2credit")
	serviceID := c.Query("service_id")
	// currency := c.Query("currency")
	trnx := c.Query("tran_id")
	count := c.Query("count")
	payment_id := c.Query("paymentid")
	// receipt := c.Query("agent_term_receipt")
	// rate := c.Query("rate")
	// fixrate := c.Query("fixrate")
	// notify := c.Query("notify")
	counts, _ := strconv.Atoi(count)
	newtrnx, _ := strconv.Atoi(trnx)
	wg.Add(counts)
	serviceIDInt, _ := strconv.Atoi(serviceID)
	fmt.Println("types --> " + types + " ---  key ---> " + key)
	switch types {

	case "precheck":

		for i := 0; i < counts; i++ {
			hashText := psID + fmt.Sprint(newtrnx) + key
			hashMustBe := GetSha512Hash(hashText)
			req := Request{
				Command:     "precheck",
				Login:       psID,
				ServiceID:   serviceIDInt,
				Phone:       number,
				RequestType: "online",
				Amount:      0.1,
				TransID:     fmt.Sprint(newtrnx),
				HashSum:     hashMustBe,
			}
			var buf bytes.Buffer
			xml.NewEncoder(&buf).Encode(req)
			bJosn, _ := json.Marshal(req)
			fmt.Println(string(bJosn))
			go SendRequestBody(&buf, fmt.Sprint(trnx), &wg)
			newtrnx++
		}

	case "payment":

		for i := 0; i < counts; i++ {
			hashText := psID + fmt.Sprint(newtrnx) + key
			hashMustBe := GetSha512Hash(hashText)
			req := Request{
				Command:     "topup",
				Login:       psID,
				ServiceID:   serviceIDInt,
				Phone:       number,
				RequestType: "topup",
				Amount:      0.1,
				TransID:     fmt.Sprint(newtrnx),
				HashSum:     hashMustBe,
			}
			var buf bytes.Buffer
			xml.NewEncoder(&buf).Encode(req)
			bJosn, _ := json.Marshal(req)
			fmt.Println(string(bJosn))
			go SendRequestBody(&buf, fmt.Sprint(trnx), &wg)
			newtrnx++
		}
	case "postcheck":
		for i := 0; i < counts; i++ {
			hashText := psID + payment_id + key
			hashMustBe := GetSha512Hash(hashText)
			req := Request{
				Command:     "postcheck",
				Login:       psID,
				ServiceID:   serviceIDInt,
				Phone:       number,
				RequestType: "online",
				Amount:      0.1,
				TransID:     fmt.Sprint(newtrnx),
				HashSum:     hashMustBe,
			}
			var buf bytes.Buffer
			xml.NewEncoder(&buf).Encode(req)
			bJosn, _ := json.Marshal(req)
			fmt.Println(string(bJosn))
			go SendRequestBody(&buf, fmt.Sprint(payment_id), &wg)
			newtrnx++
		}
	}

	wg.Wait()
	fmt.Println("avr time:", float64(sumTime.Load().(int64))/float64(counts))
	c.JSON(200, float64(sumTime.Load().(int64))/float64(counts))
	sumTime.Store(int64(0))
}
func mainRoute(c *gin.Context) {
	sumTime = atomic.Value{}
	errCount = atomic.Value{}
	sumTime.Store(int64(0))
	errCount.Store(0)
	var wg sync.WaitGroup
	psID := c.Query("ps_id")
	key := c.Query("key")
	number := c.Query("number")
	types := c.Query("type")
	amount := c.Query("amount")
	amount2credit := c.Query("amount2credit")
	serviceID := c.Query("service_id")
	currency := c.Query("currency")
	trnx := c.Query("tran_id")
	checkType := c.Query("checkType")

	count := c.Query("count")
	receipt := c.Query("agent_term_receipt")
	rate := c.Query("rate")
	fixrate := c.Query("fixrate")
	// durationStr := c.Query("duration")
	counts, _ := strconv.Atoi(count)
	// duration, _ := strconv.Atoi(durationStr)
	// timeMin := float64(0)
	// timeMax := float64(0)
	// times := []float64{}
	newtrnx, _ := strconv.Atoi(trnx)
	notify := c.Query("notify")
	payment_id := c.Query("paymentid")
	// now := time.Now()
	countCircle := 0
	// for {
	// 	if time.Since(now).Seconds() >= float64(duration) {
	// 		break
	// 	}
	countCircle++
	// wg.Add(counts)
	fmt.Println("types --> " + types + " ---  key ---> " + key)
	switch types {
	case "createpayment":
		fmt.Println("START!!!!")

		for i := 0; i < counts; i++ {
			q := url.Values{}

			data := []byte(psID + number + amount + currency + serviceID + fmt.Sprint(newtrnx))

			hash := ValidMACByte(data, []byte(key))
			q.Add("request", types)
			q.Add("ps_id", psID)
			q.Add("account", number)
			q.Add("amount", amount)
			q.Add("amount2credit", amount2credit)
			q.Add("service_id", serviceID)
			q.Add("trnx_id", fmt.Sprint(newtrnx))
			q.Add("hash", hash)
			q.Add("currency", currency)
			q.Add("notify_route", notify)

			go SendRequest(q.Encode(), fmt.Sprint(trnx), &wg)
			newtrnx++
		}

	case "precheck":

		for i := 0; i < counts; i++ {
			q := url.Values{}

			data := []byte(psID + number + amount + currency + serviceID + fmt.Sprint(newtrnx))

			hash := ValidMACByte(data, []byte(key))
			q.Add("request", types)
			q.Add("ps_id", psID)
			q.Add("account", number)
			q.Add("amount", amount)
			// q.Add("amount2credit", amount2credit)
			q.Add("service_id", serviceID)
			q.Add("trnx_id", fmt.Sprint(newtrnx))
			q.Add("hash", hash)
			q.Add("currency", currency)
			q.Add("notify_route", notify)

			go SendRequest(q.Encode(), fmt.Sprint(trnx), &wg)
			newtrnx++
		}
	case "smartCheck":

		for i := 0; i < counts; i++ {
			q := url.Values{}

			data := []byte(psID + serviceID)

			hash := ValidMACByte(data, []byte(key))
			q.Add("request", types)
			q.Add("ps_id", psID)
			q.Add("account", number)
			q.Add("amount", amount)
			q.Add("check_type", checkType)
			q.Add("grp_id", serviceID)
			q.Add("trnx_id", fmt.Sprint(newtrnx))
			q.Add("hash", hash)
			q.Add("currency", currency)
			q.Add("notify_route", notify)

			go SendRequest(q.Encode(), fmt.Sprint(trnx), &wg)
			newtrnx++
		}

	case "payment":

		for i := 0; i < counts; i++ {
			q := url.Values{}

			data := []byte(psID + number + amount + serviceID + fmt.Sprint(newtrnx))
			hash := ValidMACByte(data, []byte(key))
			fmt.Println(hash)
			q.Add("request", types)
			q.Add("ps_id", psID)
			q.Add("account", number)
			q.Add("amount", amount)
			q.Add("amount2credit", amount2credit)
			q.Add("service_id", serviceID)
			q.Add("agent_term_receipt", receipt)
			q.Add("trnx_id", fmt.Sprint(newtrnx))
			q.Add("hash", hash)
			q.Add("currency", currency)
			q.Add("notify_route", notify)
			q.Add("rate", rate)
			q.Add("fixrate", fixrate)
			go SendRequest(q.Encode(), fmt.Sprint(newtrnx), &wg)
			newtrnx++
		}

	case "getserviceslist":
		q := url.Values{}
		data := []byte(psID + fmt.Sprint(trnx))
		hash := ValidMACByte(data, []byte(key))
		fmt.Println(hash)
		q.Add("request", types)
		q.Add("ps_id", psID)
		q.Add("trnx_id", fmt.Sprint(trnx))
		q.Add("hash", hash)

		go SendRequest(q.Encode(), fmt.Sprint(trnx), &wg)

	case "getbalance":
		q := url.Values{}
		data := []byte(psID + fmt.Sprint(trnx))
		hash := ValidMACByte(data, []byte(key))
		fmt.Println(hash)
		q.Add("request", types)
		q.Add("ps_id", psID)
		q.Add("trnx_id", fmt.Sprint(trnx))
		q.Add("hash", hash)

		go SendRequest(q.Encode(), fmt.Sprint(trnx), &wg)

	case "postcheck":
		q := url.Values{}
		fmt.Println("payment_id ---> ", payment_id)
		data := []byte(psID + fmt.Sprint(trnx))
		hash := ValidMACByte(data, []byte(key))
		fmt.Println(hash)
		q.Add("request", types)
		q.Add("ps_id", psID)
		q.Add("trnx_id", fmt.Sprint(trnx))
		q.Add("hash", hash)

		go SendRequest(q.Encode(), fmt.Sprint(trnx), &wg)
	}

	// wg.Wait()
	// fmt.Println("avr time:", float64(sumTime.Load().(int64))/float64(counts))
	// if timeMin == 0 {
	// 	timeMin = float64(sumTime.Load().(int64)) / float64(counts)
	// }
	// if timeMax == 0 {
	// 	timeMax = float64(sumTime.Load().(int64)) / float64(counts)
	// }
	// if float64(sumTime.Load().(int64))/float64(counts) < timeMin {
	// 	timeMin = float64(sumTime.Load().(int64)) / float64(counts)
	// }

	// if float64(sumTime.Load().(int64))/float64(counts) > timeMax {
	// 	timeMax = float64(sumTime.Load().(int64)) / float64(counts)
	// }
	// times = append(times, float64(sumTime.Load().(int64))/float64(counts))
	// sumTime.Store(int64(0))

	// timesSum := float64(0)
	// for _, v := range times {
	// 	timesSum += v
	// }

	c.JSON(200, gin.H{
		"lastTrnID": newtrnx,
		// "errCount":   errCount.Load(),
		// "maxTimeReq": timeMax,
		// "minTimeReq": timeMin,
		// "avgTimeReq": timesSum / float64(len(times)),
	})

}

var sumTime atomic.Value
var errCount atomic.Value

func SendRequest(query string, id string, wg *sync.WaitGroup) {
	// now := time.Now()
	var response ResponseFromXml

	client := &http.Client{
		Timeout: time.Duration(20) * time.Second,
		Transport: &http.Transport{
			// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy: http.ProxyURL(nil),
		},
	}

	req, err := http.NewRequest("GET", "http://192.168.150.64:8003/", nil)

	if err != nil {
		log.Println("[retailgate request]:", err)
		return
	}

	req.URL.RawQuery = query
	log.Println("[retailgate request]:", req.URL.RawQuery)

	resp, err := client.Do(req)

	if err != nil {
		errCount.Store(errCount.Load().(int) + 1)
		log.Println("[humokmfund response]:", err)
		return
	}
	v, _ := ioutil.ReadAll(resp.Body)

	err = xml.Unmarshal(v, &response)
	if err != nil || response.Status >= 400 {
		errCount.Store(errCount.Load().(int) + 1)
	}
	log.Println("[humokmfund response]:", response, nil)
	log.Println("[humokmfund response 111111111111111111]:", string(v))
	// f, err := os.OpenFile("data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer f.Close()

	// log.SetOutput(f)
	// log.Println(id, "---", string(v))
	// sumTime.Swap(sumTime.Load().(int64) + time.Since(now).Milliseconds())
	// wg.Done()
}
func SendRequestBody(body io.Reader, id string, wg *sync.WaitGroup) {
	// now := time.Now()
	var response ResponseFromXml

	client := &http.Client{
		Timeout: time.Duration(20) * time.Second,
		Transport: &http.Transport{
			// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy: http.ProxyURL(nil),
		},
	}

	req, err := http.NewRequest("POST", "http://192.168.0.163:8860/life10/hamsoya", body)

	if err != nil {
		log.Println("[retailgate request]:", err)
		return
	}

	log.Println("[retailgate request]:", req.URL.RawQuery)

	resp, err := client.Do(req)

	if err != nil {
		log.Println("[humokmfund response]:", err)
		return
	}
	v, _ := ioutil.ReadAll(resp.Body)

	err = xml.Unmarshal(v, &response)

	log.Println("[humokmfund response]:", response, nil)
	log.Println("[humokmfund response 111111111111111111]:", string(v))
	// f, err := os.OpenFile("data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer f.Close()

	// log.SetOutput(f)
	// log.Println(id, "---", string(v))
	// sumTime.Swap(sumTime.Load().(int64) + time.Since(now).Milliseconds())
	wg.Done()
}
