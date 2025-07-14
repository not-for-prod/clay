package mwcommon

import (
	"context"

	"github.com/not-for-prod/clay/server/middlewares/mwcommon"
)

func GetLogFunc(logger interface{}) func(context.Context, string) {
	return mwcommon.GetLogFunc(logger)
}
