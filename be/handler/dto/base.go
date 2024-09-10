package dto

type CommonResp[T any] struct {
	ReturnCode       int    `json:"return_code,omitempty"`
	ReturnMessage    string `json:"return_message,omitempty"`
	SubReturnCode    int    `json:"sub_return_code,omitempty"`
	SubReturnMessage string `json:"sub_return_message,omitempty"`
	Data             T      `json:"data,omitempty"`
}
