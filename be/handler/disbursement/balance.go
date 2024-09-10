package disbursement

import (
	"context"
	"demo/constant"
	"demo/handler/dto"
	"demo/handler/util"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// all field is server construct
type QueryBalanceRequest struct {
	RequestId string `json:"request_id,omitempty" mapstructure:"request_id"`
	AppId     int    `json:"app_id,omitempty" mapstructure:"app_id"`
	PaymentId string `json:"payment_id,omitempty" mapstructure:"payment_id"`
	Time      int64  `json:"time,omitempty" mapstructure:"time"`
	Mac       string `json:"mac,omitempty" mapstructure:"mac"`
}

type QueryBalanceResponseData struct {
	Balance int64 `json:"balance,omitempty" mapstructure:"balance"`
}

type QueryBalanceResponse = dto.CommonResp[QueryBalanceResponseData]

func queryBalanceHandler(ctx context.Context, _ *QueryBalanceRequest) (*QueryBalanceResponse, error) {
	cur := time.Now()
	mac := util.HmacEncode(
		constant.Config.AppIdStr,
		constant.Config.PaymentId,
		strconv.FormatInt(cur.UnixMilli(), 10),
	)
	return util.HttpJsonReqJsonResponse[*QueryBalanceRequest, *QueryBalanceResponse](ctx, &QueryBalanceRequest{
		RequestId: uuid.NewString(),
		AppId:     constant.Config.AppId,
		PaymentId: constant.Config.PaymentId,
		Time:      cur.UnixMilli(),
		Mac:       mac,
	}, "/disbursement/balance")
}

var QueryBalanceHandler = util.ToGinHandler(queryBalanceHandler)
