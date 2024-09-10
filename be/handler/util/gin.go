package util

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ToGinHandler[T PtrTo[Req], R PtrTo[Res], Req, Res any](handler func(ctx context.Context, req T) (Res, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(Req) // new Req => construct empty struct instead of *new(T) => construct nil
		raw, _ /* should handle err */ := io.ReadAll(ctx.Request.Body)
		json.Unmarshal(raw, req)
		log.Println(string(raw))
		res, err := handler(ctx, req)
		if err != nil {
			log.Print(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		resBytes, _ := json.Marshal(res)
		ctx.Writer.Write(resBytes)
	}

}
