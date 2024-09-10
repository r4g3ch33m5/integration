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

type QueryUserRequest struct {
	Phone string `json:"phone,omitempty"` // FE passed this field

	// server construct
	AppId     int    `json:"app_id,omitempty" mapstructure:"app_id"`
	Time      int64  `json:"time,omitempty" mapstructure:"time"`
	RequestId string `json:"request_id,omitempty" mapstructure:"request_id"`
	Mac       string `json:"mac,omitempty" mapstructure:"mac"`
}

type QueryUserResponse = dto.CommonResp[QueryUserResponseData]

type QueryUserResponseData struct {
	ReferenceId string `json:"reference_id,omitempty"`
	MUId        string `json:"m_u_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

func queryUserHandler(ctx context.Context, req *QueryUserRequest) (*QueryUserResponse, error) {

	// add server field
	curTime := time.Now().UnixMilli()
	req.AppId = constant.Config.AppId
	req.Time = curTime
	req.RequestId = uuid.NewString()
	req.Mac = util.HmacEncode(constant.Config.AppIdStr, req.Phone, strconv.Itoa(int(curTime)))
	return util.HttpJsonReqJsonResponse[*QueryUserRequest, *QueryUserResponse](ctx, req, "/disbursement/user")
}

type VerifyAccountRequest struct {
	UserInfo struct {
		BankCode          string `json:"bank_code,omitempty" mapstructure:"bank_code" form:"bank_code"`
		AccountNo         string `json:"account_no,omitempty" mapstructure:"account_no" form:"account_no"`
		AccountHolderName string `json:"account_holder_name,omitempty" mapstructure:"account_holder_name" form:"account_holder_name"`
		Phone             string `json:"phone,omitempty" mapstructure:"phone" form:"phone"`
		MUId              string `json:"m_u_id,omitempty" mapstructure:"m_u_id" form:"mu_id"`
	} `mapstructure:"user_info" json:"user_info,omitempty" form:"user_info"` // FE send raw, server encrypt
	Amount int64  `json:"amount,omitempty" mapstructure:"amount" form:"amount"`
	Type   string `json:"type,omitempty" mapstructure:"type" form:"type"`
}

type ZlpVerifyUserRequest struct {
	AppId            int    `json:"app_id,omitempty" mapstructure:"app_id"`
	DisbursementType string `json:"disbursement_type,omitempty" mapstructure:"disbursement_type"`
	// receive from FE to encode into receiver info
	ReceiverInfo string `json:"receiver_info,omitempty" mapstructure:"receiver_info"`
	Amount       int64  `json:"amount,omitempty" mapstructure:"amount"`
	RedirectUrl  string `json:"redirect_url,omitempty" mapstructure:"redirect_url"` // optional
	Time         int64  `json:"time,omitempty" mapstructure:"time"`
	Mac          string `json:"mac,omitempty" mapstructure:"mac"`
}

type VerifyAccountResponse = dto.CommonResp[struct {
	MUId              string `json:"m_u_id,omitempty" mapstructure:"m_u_id"`
	AccountHolderName string `json:"account_holder_name,omitempty" mapstructure:"account_holder_name"`
	CardHolderName    string `json:"card_holder_name,omitempty" mapstructure:"card_holder_name"`
	ReformUrl         string `json:"reform_url,omitempty" mapstructure:"reform_url"`
}]

func (r *VerifyAccountRequest) ToZlpRequest() *ZlpVerifyUserRequest {
	cur := time.Now()
	// receiverInfo := util.HmacEncode(

	// )
	log.Println(r.UserInfo)
	receiver, err := json.Marshal(r.UserInfo)
	if err != nil {
		log.Println(err)
	}
	receiverInfo := util.RsaPubEncode(receiver)
	return &ZlpVerifyUserRequest{
		AppId:            constant.Config.AppId,
		DisbursementType: r.Type, // hardcode
		Time:             cur.UnixMilli(),
		ReceiverInfo:     receiverInfo,
		Amount:           r.Amount,
		// redirect url
		Mac: util.HmacEncode(
			constant.Config.AppIdStr,
			r.Type,
			receiverInfo,
			strconv.Itoa(int(r.Amount)),
			strconv.Itoa(int(cur.UnixMilli())),
		),
	}
}

func verifyAcountHandler(ctx context.Context, req *VerifyAccountRequest) (*VerifyAccountResponse, error) {
	return util.HttpJsonReqJsonResponse[*ZlpVerifyUserRequest, *VerifyAccountResponse](ctx, req.ToZlpRequest(), "/disbursement/verify-account")
}

type ListBankCodeResp = dto.CommonResp[struct {
	BankList []struct {
		BankCode  string `json:"bank_code,omitempty" mapstructure:"bank_code" form:"bank_code"`
		ShortName string `json:"short_name,omitempty" mapstructure:"short_name" form:"short_name"`
		Name      string `json:"name,omitempty" mapstructure:"name" form:"name"`
		LogoUrl   string `json:"logo_url,omitempty" mapstructure:"logo_url" form:"logo_url"`
	} `json:"bank_list,omitempty" mapstructure:"bank_list" form:"bank_list"`
}]

func listBankCodeHandler(ctx context.Context, _ *any) (*ListBankCodeResp, error) {
	cur := time.Now()
	return util.HttpJsonReqJsonResponse[*struct {
		AppId int    `json:"app_id,omitempty" mapstructure:"app_id" form:"app_id"`
		Time  int64  `json:"time,omitempty" mapstructure:"time" form:"time"`
		Mac   string `json:"mac,omitempty" mapstructure:"mac" form:"mac"`
	}, *ListBankCodeResp](
		ctx,
		&struct {
			AppId int    `json:"app_id,omitempty" mapstructure:"app_id" form:"app_id"`
			Time  int64  `json:"time,omitempty" mapstructure:"time" form:"time"`
			Mac   string `json:"mac,omitempty" mapstructure:"mac" form:"mac"`
		}{
			AppId: constant.Config.AppId,
			Time:  cur.UnixMilli(),
			Mac:   util.HmacEncode(constant.Config.AppIdStr, strconv.Itoa(int(cur.UnixMilli()))),
		},
		"/disbursement/get-bank-code",
	)
}

var (
	ListBankCodeHandler = util.ToGinHandler(listBankCodeHandler)
	QueryUserHandler    = util.ToGinHandler(queryUserHandler)
	VerifyUserHandler   = util.ToGinHandler(verifyAcountHandler)
)
