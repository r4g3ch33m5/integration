package handler

import (
	"context"
	"demo/constant"
	"demo/handler/util"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Order struct {
	AppTransId string `json:"app_trans_id,omitempty" mapstructure:"app_trans_id" form:"app_trans_id"`
	Status     int    `json:"status,omitempty" mapstructure:"status" form:"status"`
	OrderLink  string `json:"order_link,omitempty" mapstructure:"order_link" form:"order_link"`
	QrCode     string `json:"qr_code,omitempty" mapstructure:"qr_code" form:"qr_code"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Order)
	upgrader  = websocket.Upgrader{}
	mu        sync.Mutex
)

// pointer to
type RequestType[T any] interface {
	*T
}

// pointer to
type ResponseType[T any] interface {
	*T
}

type object map[string]any

type CreateOrderReq struct {
	Amount int `json:"amount,omitempty"`
}
type CreateOrderRes struct {
	AppTransId string `json:"app_trans_id,omitempty"`
	OrderLink  string `json:"order_link,omitempty"`
	QrCode     string `json:"qr_code,omitempty"`
}

func handleCreate(ctx context.Context, request *CreateOrderReq) (*CreateOrderRes, error) {
	rand.Seed(time.Now().UnixNano())
	var (
		now        = time.Now()
		transID    = rand.Intn(10000000000) // Generate random trans id
		appTransID = fmt.Sprintf("%02d%02d%02d_%v", now.Year()%100, int(now.Month()), now.Day(), transID)
	)
	embedData := object{}
	embedData["redirecturl"] = fmt.Sprintf("%v/fe", constant.Config.CallBackUrl)
	items, _ := json.Marshal([]object{})
	// request data
	params := make(url.Values)
	params.Add("app_id", constant.Config.AppIdStr)
	params.Add("app_user", "thinhlth")
	params.Add("app_trans_id", appTransID)                         // translation missing: vi.docs.shared.sample_code.comments.app_trans_id
	params.Add("app_time", strconv.FormatInt(now.UnixMilli(), 10)) // miliseconds
	params.Add("amount", strconv.Itoa(request.Amount))
	params.Add("item", string(items))
	params.Add("description", "Lazada - Payment for the order #"+strconv.Itoa(transID))
	embedBytes, _ := json.Marshal(embedData)
	params.Add("embed_data", string(embedBytes))
	params.Add("bankcode", "zalopayapp")
	params.Add("callback_url", fmt.Sprintf("%v/%v", constant.Config.CallBackUrl, "api/callback"))
	// appid|app_trans_id|appuser|amount|apptime|embeddata|item
	fmt.Println(params)
	mac := util.HmacEncode(params.Get("app_id"), params.Get("app_trans_id"), params.Get("app_user"),
		params.Get("amount"), params.Get("app_time"), params.Get("embed_data"), params.Get("item"))
	params.Add("mac", mac)

	res, err := http.PostForm("https://sb-openapi.zalopay.vn/v2/create", params)

	// parse response
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res.StatusCode)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var result map[string]interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal(err)
	}
	// log.Print(params)
	for k, v := range result {
		log.Printf("create order %s = %+v", k, v)
	}

	Db[appTransID] = Order{
		AppTransId: appTransID,
		OrderLink:  result["order_url"].(string),
		QrCode:     result["qr_code"].(string),
		Status:     0, // init
	}

	log.Println(Db[appTransID])

	return &CreateOrderRes{
		AppTransId: appTransID,
		OrderLink:  result["order_url"].(string),
		QrCode:     result["qr_code"].(string),
	}, nil
}

var CreateOrderHandler = util.ToGinHandler(handleCreate)
