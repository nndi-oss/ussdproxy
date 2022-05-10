package server

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type healthcheck struct {
	Status string `json:"status"`
}

func (h healthcheck) JSON() ([]byte, error) {
	return json.Marshal(h)
}
func healthy() healthcheck {
	return healthcheck{
		Status: "HEALTHY",
	}
}
func unhealthy() healthcheck {
	return healthcheck{
		Status: "UNHEALTHY",
	}
}

func (s *UssdProxyServer) healthcheckHandler(ctx *fasthttp.RequestCtx) {
	data, err := healthy().JSON()
	if err != nil {
		panic(err)
	}
	ctx.Write(data)
}
