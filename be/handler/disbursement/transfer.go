package disbursement

import (
	"context"
	"demo/constant"
	"demo/handler/dto"
	"demo/handler/util"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type TransferFundRequestServerConstruct struct {
	AppId            int    `json:"app_id,omitempty" mapstructure:"app_id"`
	PaymentId        string `json:"payment_id,omitempty" mapstructure:"payment_id"`
	PartnerOrderId   string `json:"partner_order_id,omitempty" mapstructure:"partner_order_id"` // uuid.random
	DisbursementType string `json:"disbursement_type,omitempty" mapstructure:"disbursement_type"`
	Description      string `json:"description,omitempty" mapstructure:"description"`               // uuid.random
	PartnerEmbedData string `json:"partner_embed_data,omitempty" mapstructure:"partner_embed_data"` // hardcode
	ExtraInfo        string `json:"extra_info,omitempty" mapstructure:"extra_info"`                 // hardcode "{}"
	Time             int64  `json:"time,omitempty" mapstructure:"time"`
	Mac              string `json:"mac,omitempty" mapstructure:"mac"`
	ReceiverInfo     string `json:"receiver_info,omitempty" mapstructure:"receiver_info"` // encrypt from user_info
}

type TransferFundRequest struct {
	UserInfo *struct {
		BankCode          string `json:"bank_code,omitempty" mapstructure:"bank_code" form:"bank_code"`
		AccountNo         string `json:"account_no,omitempty" mapstructure:"account_no" form:"account_no"`
		AccountHolderName string `json:"account_holder_name,omitempty" mapstructure:"account_holder_name" form:"account_holder_name"`
		CardNo            string `json:"card_no,omitempty" mapstructure:"card_no" form:"card_no"`
		CardHolderName    string `json:"card_holder_name,omitempty" mapstructure:"card_holder_name" form:"card_holder_name"`
		Phone             string `json:"phone,omitempty" mapstructure:"phone" form:"phone"`
		MUId              string `json:"m_u_id,omitempty" mapstructure:"m_u_id" form:"mu_id"`
	} `json:"user_info,omitempty" mapstructure:"user_info" form:"user_info"`
	Amount int64  `json:"amount,omitempty" mapstructure:"amount" form:"amount"`
	Type   string `json:"type,omitempty" mapstructure:"type" form:"type"`
	TransferFundRequestServerConstruct
}

// construct mac
func (r *TransferFundRequest) FinalizeRequest() {
	raw, _ := json.Marshal(r.UserInfo)
	log.Println("receiver info:\n", string(raw))
	r.ReceiverInfo = util.RsaPubEncode(raw)
	r.UserInfo = nil
	// mac must be last
	r.Mac = util.HmacEncode(
		constant.Config.AppIdStr,
		r.PaymentId,
		r.PartnerOrderId,
		r.DisbursementType,
		r.ReceiverInfo,
		strconv.Itoa(int(r.Amount)),
		r.Description,
		r.PartnerEmbedData,
		r.ExtraInfo,
		strconv.Itoa(int(r.Time)),
	)

}

type TransferFundResponse = dto.CommonResp[struct {
	OrderId           string `json:"order_id,omitempty" mapstructure:"order_id"`
	DisbursementType  string `json:"disbursement_type,omitempty" mapstructure:"disbursement_type"`
	MUId              string `json:"m_u_id,omitempty" mapstructure:"m_u_id"`
	Phone             string `json:"phone,omitempty" mapstructure:"phone"`
	BankCode          string `json:"bank_code,omitempty" mapstructure:"bank_code"`
	AccountNo         string `json:"account_no,omitempty" mapstructure:"account_no"`
	AccountHolderName string `json:"account_holder_name,omitempty" mapstructure:"account_holder_name"`
	CardNo            string `json:"card_no,omitempty" mapstructure:"card_no"`
	CardHolderName    string `json:"card_holder_name,omitempty" mapstructure:"card_holder_name"`
	Status            int    `json:"status,omitempty" mapstructure:"status"`
	Amount            int64  `json:"amount,omitempty" mapstructure:"amount"`
	PartnerFee        int64  `json:"partner_fee,omitempty" mapstructure:"partner_fee"`
	ZlpFee            int64  `json:"zlp_fee,omitempty" mapstructure:"zlp_fee"`
	ServerTime        int64  `json:"server_time,omitempty" mapstructure:"server_time"`
}]

func transferFundHandler(ctx context.Context, req *TransferFundRequest) (*TransferFundResponse, error) {
	cur := time.Now()
	orderId := uuid.NewString()
	log.Println(req.UserInfo)
	req.TransferFundRequestServerConstruct = TransferFundRequestServerConstruct{
		AppId:            constant.Config.AppId,
		PaymentId:        constant.Config.PaymentId,
		PartnerOrderId:   orderId,
		DisbursementType: req.Type,
		Description:      uuid.NewString(),
		PartnerEmbedData: "{}",
		ExtraInfo:        "{}",
		Time:             cur.UnixMilli(),
	}
	req.FinalizeRequest()

	resp, err := util.HttpJsonReqJsonResponse[*TransferFundRequest, *TransferFundResponse](ctx, req, "/disbursement/transfer-fund")
	if err != nil {
		return nil, err
	}
	resp.Data.OrderId = orderId
	return resp, nil
}


var TransferFundHandler = util.ToGinHandler(transferFundHandler)
