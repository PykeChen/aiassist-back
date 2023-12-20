package main

import (
	"github.com/gin-gonic/gin"
	"hy_golang_sdk/pkg/plog"
	"hy_golang_sdk/qrcode"
)

func main() {

	plog.InitFileLogger(".", "logic")
	ginInstance := gin.Default()
	common := ginInstance.Group("common")
	common.GET("/qrcode", qrcode.GenerateQR)

	err := ginInstance.Run(":30034")
	if err != nil {
		return
	}
}
