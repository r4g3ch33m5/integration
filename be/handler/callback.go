package handler

import (
	"context"
	"demo/constant"
	"demo/handler/util"
	"encoding/json"
	"log"

	"github.com/zpmep/hmacutil"
)

type CallbackRequest struct {
	Data         string
	Mac          string
	CallbackType int
}

type CallbackResponse struct {
	ReturnCode    int    `json:"return_code,omitempty"`
	ReturnMessage string `json:"return_message,omitempty"`
}

func handleOrderUpdateCallback(ctx context.Context, req *CallbackRequest) (*CallbackResponse, error) {
	log.Println("receive msg", req)
	mac := hmacutil.HexStringEncode(hmacutil.SHA256, constant.Config.Key2, req.Data)
	// kiểm tra callback hợp lệ (đến từ ZaloPay server)
	if mac != req.Mac {
		// callback không hợp lệ
		return &CallbackResponse{
			ReturnCode:    -1,
			ReturnMessage: "mac not equal",
		}, nil
	}

	// merchant cập nhật trạng thái cho đơn hàng
	var dataJSON map[string]interface{}
	json.Unmarshal([]byte(req.Data), &dataJSON)
	log.Println("update order's status = success where app_trans_id =", dataJSON["app_trans_id"])

	broadcast <- Order{
		AppTransId: dataJSON["app_trans_id"].(string),
		Status:     1, // success
	}

	return &CallbackResponse{ // thanh toán thành công
		ReturnCode:    1,
		ReturnMessage: "success",
	}, nil
}

var OrderUpdateCallbackHandler = util.ToGinHandler(handleOrderUpdateCallback)
