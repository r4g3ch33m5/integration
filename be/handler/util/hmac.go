package util

import (
	"crypto/rand"
	"crypto/rsa"
	"demo/constant"
	"encoding/base64"
	"log"
	"strings"

	"github.com/zpmep/hmacutil"
)

func HmacEncode(data ...string) string {
	log.Println("hmac input:", strings.Join(data, "|"))
	mac := hmacutil.HexStringEncode(hmacutil.SHA256, constant.Config.Key1, strings.Join(data, "|"))
	log.Println("hmac output", mac)
	return mac
}

func RsaPubEncode(data []byte) string {
	
	out, err := rsa.EncryptPKCS1v15(rand.Reader, constant.Config.RsaPubKey, data)
	if err != nil {	
		log.Println(err)
	}
	return base64.StdEncoding.EncodeToString(out)
}
