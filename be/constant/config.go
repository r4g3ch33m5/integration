package constant

import "crypto/rsa"

type ConfigType struct {
	CallBackUrl   string          `mapstructure:"call_back_url,omitempty" json:"call_back_url,omitempty"`
	ZlpApiUrl     string          `json:"zlp_api_url,omitempty" mapstructure:"zlp_api_url"`
	AppIdStr      string          `json:"app_id_str,omitempty" mapstructure:"app_id_str"`
	AppId         int             `json:"app_id,omitempty" mapstructure:"app_id"`
	Key1          string          `json:"key_1,omitempty" mapstructure:"key_1"`
	Key2          string          `json:"key_2,omitempty" mapstructure:"key_2"`
	PaymentId     string          `json:"payment_id,omitempty" mapstructure:"payment_id"`
	PrivateKey    string          `json:"private_key,omitempty" mapstructure:"private_key"`
	RsaPrivateKey *rsa.PrivateKey `json:"rsa_public_key,omitempty" mapstructure:"rsa_public_key"` // loadconfig will load it when program start
	PubKey        string          `json:"pub_key,omitempty" mapstructure:"pub_key"`
	RsaPubKey     *rsa.PublicKey
}

var Config = ConfigType{}
