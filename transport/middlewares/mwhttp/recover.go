package mwhttp

import (
	"github.com/not-for-prod/clay/server/middlewares/mwhttp"
)

// Recover recovers HTTP server from handlers' panics.
func Recover(logger interface{}) Middleware {
	return mwhttp.Recover(logger)
}
