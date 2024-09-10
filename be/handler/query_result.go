package handler

import (
	"bytes"
	"context"
	"demo/constant"
	"demo/handler/util"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/zpmep/hmacutil"
)

type OrderResultReq struct {
	AppTransId    string `json:"app_trans_id,omitempty"`
	CurrentStatus int    `json:"current_status,omitempty"`
}

type OrderResultResp struct {
	AppTransId string `json:"app_trans_id,omitempty"`
	Status     int    `json:"status,omitempty"`
	OrderLink  string `json:"order_link,omitempty"`
	QrCode     string `json:"qr_code,omitempty"`
}

func queryResult(ctx context.Context, req *OrderResultReq) (*OrderResultResp, error) {
	order, isFound := Db[req.AppTransId]
	fmt.Println(req.AppTransId)
	if !isFound {
		return nil, errors.New("not found")
	}

	// if order.Status != req.CurrentStatus { //TODO: Check status tu DB cua minh
	// 	return &OrderResultResp{AppTransId: order.AppTransId, Status: order.Status}, nil
	// }

	data := fmt.Sprintf("%v|%s|%s", constant.Config.AppIdStr, req.AppTransId, constant.Config.Key1) // appid|apptransid|key1
	params := map[string]interface{}{
		"app_id":       constant.Config.AppIdStr,
		"app_trans_id": req.AppTransId,
		"mac":          hmacutil.HexStringEncode(hmacutil.SHA256, constant.Config.Key1, data),
	}

	jsonStr, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post("https://sb-openapi.zalopay.vn/v2/query", "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var result map[string]any

	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal(err)
	}
	log.Println(string(body))
	//TODO: Chi update khi order status chua phai la final status
	order = Order{
		AppTransId: order.AppTransId,
		Status:     int(result["return_code"].(float64)),
	}
	Db[order.AppTransId] = order

	return &OrderResultResp{
		AppTransId: order.AppTransId,
		Status:     order.Status,
		OrderLink:  order.OrderLink,
		QrCode:     order.QrCode,
	}, nil
}

var QueryResultHandler = util.ToGinHandler(queryResult)
