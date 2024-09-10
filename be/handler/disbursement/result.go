package disbursement

import (
	"context"
	"demo/constant"
	"demo/handler/dto"
	"demo/handler/util"
	"strconv"
	"time"
)

type QueryResultRequest struct {
	PartnerOrderId                    string `json:"partner_order_id,omitempty" mapstructure:"partner_order_id"`
	QueryResultRequestServerConstruct `mapstructure:",squash"`
}

type QueryResultRequestServerConstruct struct {
	AppId int    `json:"app_id,omitempty" mapstructure:"app_id"`
	Time  int64  `json:"time,omitempty" mapstructure:"time"`
	Mac   string `json:"mac,omitempty" mapstructure:"mac"`
}

type QueryResultResponseData struct {
	OrderId           string `json:"order_id,omitempty" mapstructure:"order_id"`
	DisbursementType  string `json:"disbursement_type,omitempty" mapstructure:"disbursement_type"` // hardcode
	MUId              string `json:"m_u_id,omitempty" mapstructure:"m_u_id"`
	Phone             string `json:"phone,omitempty" mapstructure:"phone"`
	BankCode          string `json:"bank_code,omitempty" mapstructure:"bank_code"`
	AccountNo         string `json:"account_no,omitempty" mapstructure:"account_no"`
	AccountHolderName string `json:"account_holder_name,omitempty" mapstructure:"account_holder_name"`
	CardNo            string `json:"card_no,omitempty" mapstructure:"card_no"`
	CardHolderName    string `json:"card_holder_name,omitempty" mapstructure:"card_holder_name"`
	Status            int    `json:"status,omitempty" mapstructure:"status"`
	Amount            int64  `json:"amount,omitempty" mapstructure:"amount"`
	ZpTransId         string `json:"zp_trans_id,omitempty" mapstructure:"zp_trans_id"`
	PartnerFee        int64  `json:"partner_fee,omitempty" mapstructure:"partner_fee"`
	ZlpFee            int64  `json:"zlp_fee,omitempty" mapstructure:"zlp_fee"`
	ServerTime        int64  `json:"server_time,omitempty" mapstructure:"server_time"`
}

type QueryResultResponse = dto.CommonResp[QueryResultResponseData]

func queryResultHandler(ctx context.Context, req *QueryResultRequest) (*QueryResultResponse, error) {
	cur := time.Now().UnixMilli()
	mac := util.HmacEncode(constant.Config.AppIdStr, req.PartnerOrderId, strconv.Itoa(int(cur)))
	req.QueryResultRequestServerConstruct = QueryResultRequestServerConstruct{
		AppId: constant.Config.AppId,
		Time:  cur,
		Mac:   mac,
	}
	return util.HttpJsonReqJsonResponse[*QueryResultRequest, *QueryResultResponse](
		ctx, req, "/disbursement/query-txn",
	)
}

var QueryResultHandler = util.ToGinHandler(queryResultHandler)
