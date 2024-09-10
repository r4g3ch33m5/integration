package disbursement

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"demo/constant"
	"demo/handler/dto"
	"demo/handler/util"
	"encoding/base64"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type TopupRequest struct {
	// FE pass
	Amount int64 `json:"amount,omitempty" mapstructure:"amount"`
	// from query user
	MUId        string `json:"m_u_id,omitempty" mapstructure:"m_u_id"`
	ReferenceId string `json:"reference_id,omitempty" mapstructure:"reference_id"`

	TopupRequestServerConstruct `mapstructure:",squash"`
}

// server construct
type TopupRequestServerConstruct struct {
	AppId            int    `json:"app_id,omitempty" mapstructure:"app_id"`
	PaymentId        string `json:"payment_id,omitempty" mapstructure:"payment_id"`
	PartnerOrderId   string `json:"partner_order_id,omitempty" mapstructure:"partner_order_id"`
	PartnerEmbedData string `json:"partner_embed_data,omitempty" mapstructure:"partner_embed_data"` // hardcode
	Time             int64  `json:"time,omitempty" mapstructure:"time"`
	Sig              string `json:"sig,omitempty" mapstructure:"sig"`
	Description      string `json:"description,omitempty" mapstructure:"description"` // hardcode
	ExtraInfo        string `json:"extra_info,omitempty" mapstructure:"extra_info"`   // hardcode
}

type TopupResponseData struct {
	OrderId     string `json:"order_id,omitempty" mapstructure:"order_id"`
	Status      int    `json:"status,omitempty" mapstructure:"status"`
	MUId        string `json:"mu_id,omitempty" mapstructure:"mu_id"`
	Phone       string `json:"phone,omitempty" mapstructure:"phone"`
	Amount      int64  `json:"amount,omitempty" mapstructure:"amount"`
	Description string `json:"description,omitempty" mapstructure:"description"`
	PartnerFee  int64  `json:"partner_fee,omitempty" mapstructure:"partner_fee"`
	ZlpFee      int64  `json:"zlp_fee,omitempty" mapstructure:"zlp_fee"`
	ExtraInfo   string `json:"extra_info,omitempty" mapstructure:"extra_info"`
	Time        int64  `json:"time,omitempty" mapstructure:"time"`
	UpgradeUrl  string `json:"upgrade_url,omitempty" mapstructure:"upgrade_url"`
}

type TopupResponse = dto.CommonResp[TopupResponseData]

func createTopupHandler(ctx context.Context, req *TopupRequest) (*TopupResponse, error) {
	cur := time.Now().UnixMilli()
	partnerEmbedData := map[string]any{
		"merchant_wallet_id": "336A321842",
	} // empty json hardcode
	partnerEmbedDataStr, _ := json.Marshal(partnerEmbedData)
	description := uuid.NewString()
	extraInfo := "{}" // empty json hardcode
	orderId := uuid.NewString()

	msgHash := sha256.New()
	log.Println(util.HmacEncode(
		constant.Config.AppIdStr,
		constant.Config.PaymentId,
		orderId,
		req.MUId,
		strconv.Itoa(int(req.Amount)),
		description,
		string(partnerEmbedDataStr),
		extraInfo,
		strconv.Itoa(int(cur)),
	))
	_, err := msgHash.Write(
		[]byte(util.HmacEncode(
			constant.Config.AppIdStr,
			constant.Config.PaymentId,
			orderId,
			req.MUId,
			strconv.Itoa(int(req.Amount)),
			description,
			string(partnerEmbedDataStr),
			extraInfo,
			strconv.Itoa(int(cur)),
		)))
	if err != nil {
		return nil, err
	}
	sig, err := rsa.SignPSS(rand.Reader, constant.Config.RsaPrivateKey, crypto.SHA256, msgHash.Sum(nil), nil)
	if err != nil {
		return nil, err
	}
	log.Println(string(sig))
	log.Println(base64.StdEncoding.EncodeToString(sig))
	req.TopupRequestServerConstruct = TopupRequestServerConstruct{
		AppId:            constant.Config.AppId,
		PaymentId:        constant.Config.PaymentId,
		PartnerOrderId:   orderId,
		PartnerEmbedData: string(partnerEmbedDataStr),
		Description:      description,
		ExtraInfo:        extraInfo,
		Time:             cur,
		Sig:              base64.StdEncoding.EncodeToString(sig),
	}

	return util.HttpJsonReqJsonResponse[*TopupRequest, *TopupResponse](ctx, req, "/disbursement/topup")
}

var CreateTopUpHandler = util.ToGinHandler(createTopupHandler)
