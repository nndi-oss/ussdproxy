package server

import "github.com/valyala/fasthttp"

func (s *UssdProxyServer) notImplementedHandler(ctx *fasthttp.RequestCtx) {
	data, err := unhealthy().JSON()
	if err != nil {
		panic(err)
	}
	ctx.Write(data)
}
