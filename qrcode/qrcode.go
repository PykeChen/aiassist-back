package qrcode

import (
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"net/http"
)

func GenerateQR(c *gin.Context) {
	// 从请求参数中获取URL或文本
	text := c.Query("text")

	// 生成QR图片
	qr, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// 设置响应头
	c.Header("Content-Type", "image/png")

	// 将QR图片字节流写入响应
	c.Data(http.StatusOK, "image/png", qr)
}

func main() {
	// 创建Gin实例
	r := gin.Default()

	// 设置路由
	r.GET("/qrcode", GenerateQR)

	// 启动HTTP服务
	r.Run(":8080")
}
