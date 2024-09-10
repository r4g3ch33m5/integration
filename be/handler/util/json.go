package util

import (
	"bytes"
	"context"
	"demo/constant"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type PtrTo[T any] interface {
	*T
}

// uri must have prefix /
func HttpJsonReqJsonResponse[Req PtrTo[Request], Res PtrTo[Response], Request, Response any](
	_ context.Context,
	req Req,
	uri string,
) (Res, error) {
	jsonStr, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	log.Println(string(jsonStr))
	url := fmt.Sprintf("%v%v", constant.Config.ZlpApiUrl, uri)
	log.Println(url)
	rawRes, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	resp := new(Response)
	rawBytes, _ := io.ReadAll(rawRes.Body)
	log.Println(string(rawBytes))
	return resp, json.Unmarshal(rawBytes, resp)
}
