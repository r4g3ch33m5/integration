package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"demo/constant"
	"demo/handler"
	"demo/handler/disbursement"
	"demo/handler/util"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println(c.Request.URL)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3001")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func LoadConfig() {
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
	fmt.Println(viper.Unmarshal(&constant.Config))
	// convert to string
	constant.Config.AppIdStr = strconv.Itoa(constant.Config.AppId)
	// key instant init
	f, err := os.Open("./key.pem")
	fmt.Println(err)
	defer f.Close()
	raw, _ := io.ReadAll(f)
	privateKey, err := toPrivatekey(raw)
	var jsonRaw map[string]string = map[string]string{"pub_key": string(raw)}
	js, _ := json.MarshalIndent(jsonRaw, "", "\t")
	log.Println(string(js))
	constant.Config.RsaPrivateKey = privateKey

	pubF, err := os.Open("./key.pub.pem")
	fmt.Println(err)

	defer pubF.Close()
	raw, _ = io.ReadAll(pubF)
	constant.Config.RsaPubKey, err = toPublicKey(raw)
	fmt.Println(err)
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.Use(CORSMiddleware())

	// r.Any("/api/*", func(ctx *gin.Context) {
	// 	fmt.Println(ctx.Request.URL)
	// })
	// r.Use(cors.Default())
	// r.Any("/", func(ctx *gin.Context) { log.Println(ctx.Request.URL) })
	apiGroup := r.Group("/api")
	apiGroup.POST("/callback", handler.OrderUpdateCallbackHandler)
	r.GET("/ws", func(ctx *gin.Context) {
		handler.HandleWebSocketConnections(ctx.Writer, ctx.Request)
	})

	// order handler
	orderGroup := apiGroup.Group("/order")
	orderGroup.POST("/create", handler.CreateOrderHandler)
	orderGroup.POST("/result", handler.QueryResultHandler)

	disbursementGroup := apiGroup.Group("/disbursement")
	disbursementGroup.POST("/user", disbursement.QueryUserHandler)
	disbursementGroup.POST("/get-bank-code", disbursement.ListBankCodeHandler)
	disbursementGroup.POST("/verify-user", disbursement.VerifyUserHandler)
	disbursementGroup.POST("/balance", disbursement.QueryBalanceHandler)
	disbursementGroup.POST("/query-txn", disbursement.QueryResultHandler)
	disbursementGroup.POST("/topup", disbursement.CreateTopUpHandler)
	disbursementGroup.POST("/transfer-fund", disbursement.TransferFundHandler)

	apiGroup.POST("/encrypt_rsa", func(ctx *gin.Context) {
		var req struct {
			MUId              string `json:"m_u_id,omitempty" mapstructure:"mu_id" form:"mu_id"`
			BankCode          string `json:"bank_code,omitempty" mapstructure:"bank_code" form:"bank_code"`
			AccountNo         string `json:"account_no,omitempty" mapstructure:"account_no" form:"account_no"`
			AccountHolderName string `json:"account_holder_name,omitempty" mapstructure:"account_holder_name" form:"account_holder_name"`
		}

		tmp, _ := ioutil.ReadAll(ctx.Request.Body)
		json.Unmarshal(tmp, &req)
		inp, _ := json.Marshal(req)
		fmt.Println(string(inp))
		out := util.RsaPubEncode(inp)

		resp, err := json.Marshal(map[string]string{"receiver_info": out})
		fmt.Println(string(resp), err)
		ctx.Writer.Write(resp)
	})

	apiGroup.POST("/sign", func(ctx *gin.Context) {
		var req struct {
			Message string `json:"message" form:"message"`
		}
		raw, _ := ioutil.ReadAll(ctx.Request.Body)
		json.Unmarshal(raw, &req)
		fmt.Println(req)
		digest := sha256.Sum256([]byte(req.Message))

		signature, signErr := rsa.SignPKCS1v15(rand.Reader, constant.Config.RsaPrivateKey, crypto.SHA256, digest[:])

		if signErr != nil {
			fmt.Println("Could not sign message:%s", signErr.Error())
		}

		// just to check that we can survive to and from b64
		b64sig := base64.StdEncoding.EncodeToString(signature)
		ctx.Writer.Write([]byte(b64sig))
		fmt.Println(b64sig, req.Message)
		decodedSignature, _ := base64.StdEncoding.DecodeString(b64sig)
		// verify part

		verifyErr := rsa.VerifyPKCS1v15(constant.Config.RsaPubKey, crypto.SHA256, digest[:], decodedSignature)

		if verifyErr != nil {
			fmt.Println("Verification failed:", verifyErr)
		}
	})

	apiGroup.POST("/check_sig", func(ctx *gin.Context) {
		var req struct {
			Mac string `json:"mac" form:"mac"`
			Sig string `json:"sig" form:"sig"`
		}
		ctx.ShouldBind(&req)
		constant.Config.RsaPrivateKey.Public()
		decodedSignature, _ := base64.StdEncoding.DecodeString(req.Sig)
		digest := sha256.Sum256([]byte(req.Mac))
		// verify part
		fmt.Println(req.Sig, req.Mac)
		verifyErr := rsa.VerifyPKCS1v15(constant.Config.RsaPubKey, crypto.MD5, digest[:], decodedSignature)

		if verifyErr != nil {
			fmt.Println("Verification failed:", verifyErr)
			ctx.Writer.Write([]byte(verifyErr.Error()))
		}
	})
	// go handler.HandleMessages()

	apiGroup.POST("/decrypt_rsa", func(ctx *gin.Context) {
		var req struct {
			Message string `json:"message" form:"message"`
		}

		ctx.ShouldBind(&req)
		cipher, _ := base64.StdEncoding.DecodeString(req.Message)
		fmt.Println("req", string(cipher))
		originalData, err := rsa.DecryptPKCS1v15(rand.Reader, constant.Config.RsaPrivateKey, []byte(req.Message))
		fmt.Println(originalData)
		fmt.Println(err)
		if err != nil {
			ctx.Writer.Write([]byte(err.Error()))
		} else {
			ctx.Writer.Write(originalData)
		}

	})

	apiGroup.POST("/check_sig_base642hex", func(ctx *gin.Context) {
		var req struct {
			Message string `json:"message" form:"message"`
			Sig     string `json:"sig" form:"sig"`
		}

		ctx.ShouldBind(&req)
		raw, _ := base64.StdEncoding.DecodeString(req.Message)
		hexString := hex.EncodeToString(raw)
		fmt.Println(hexString)
		hashed := sha256.Sum256(raw)
		err := rsa.VerifyPKCS1v15(&constant.Config.RsaPrivateKey.PublicKey, crypto.SHA256, hashed[:], raw)
		fmt.Println(err)
	})

	apiGroup.POST("/check_sig_hex2base64", func(ctx *gin.Context) {
		var req struct {
			Message string `json:"message" form:"message"`
			Sig     string `json:"sig" form:"sig"`
		}
		ctx.ShouldBind(&req)
		fmt.Println(req)
		raw, _ := hex.DecodeString(req.Message)
		hashed := sha256.Sum256(raw)
		sig, _ := hex.DecodeString(req.Sig)
		err := rsa.VerifyPKCS1v15(
			&constant.Config.RsaPrivateKey.PublicKey,
			crypto.SHA256,
			hashed[:],
			sig,
		)
		fmt.Println("verify with mac decode" ,err)
		// verify without mac decode 
		hashed = sha256.Sum256([]byte(req.Message))
		err = rsa.VerifyPKCS1v15(
			&constant.Config.RsaPrivateKey.PublicKey,
			crypto.SHA256,
			hashed[:],
			sig,
		)
		fmt.Println("verify without mac decode", err)
	})

	return r
}

func main() {
	LoadConfig()
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

func toPrivatekey(privateKeyPem []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPem)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	iKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return iKey.(*rsa.PrivateKey), err

}

func toPublicKey(keyPem []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(keyPem)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	iKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return iKey.(*rsa.PublicKey), err

}
