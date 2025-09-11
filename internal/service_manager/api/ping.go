package api

import "github.com/fox-gonic/fox"

func (api *Api) Ping(c *fox.Context) string {
	return "pong"
}
