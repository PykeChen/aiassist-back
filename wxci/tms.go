package wxci

import (
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/profile"
	ims "github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/ims/v20201229"
)

func WxciWithImage(c *gin.Context) {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。密钥可前往官网控制台 https://console.tencentcloud.com/capi 进行获取
	imgUrl := c.Query("img")
	result := wxciImageVeify(imgUrl)
	var err error
	_, err = c.Writer.Write([]byte(result))
	if err != nil {
		return
	}
}

func wxciImageVeify(imgUrl string) string {
	credential := common.NewCredential(
		"IKIDibi8fK3wnbM7xgcRJ4kLvT0XwCGrIcWH",
		"iEdu13yzvkSy9DB3AINbmrsqjfhhaRHH",
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ims.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := ims.NewClient(credential, "ap-singapore", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := ims.NewImageModerationRequest()

	request.FileUrl = common.StringPtr(imgUrl)

	// 返回的resp是一个ImageModerationResponse的实例，与请求对象对应
	response, err := client.ImageModeration(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return fmt.Sprintf("An API error has returned: %s", err)
	}
	if err != nil {
		panic(err)
	}
	// 输出json格式的字符串回包
	fmt.Printf("%s", response.ToJsonString())
	return response.ToJsonString()
}
