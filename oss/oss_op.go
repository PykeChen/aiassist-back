package main

import (
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"hy_golang_sdk/mail_fetch"
	"hy_golang_sdk/pkg/plog"
	"net/url"
	"strings"
)

var (
	yourRegion          = "oss-cn-beijing.aliyuncs.com"
	yourAccessKeyId     = ""
	yourAccessKeySecret = ""
)

func main() {
	plog.InitFileLogger(".", "logic")
	// 定时任务去获取oss违规文件
	queryQQMailEvery2h()

	//创建一个server提供用于删除oss文件的
	gin := gin.Default()
	ossG := gin.Group("oss")
	ossG.POST("del", ossDel)
	err := gin.Run(":30030")
	if err != nil {
		return
	}
}

func queryQQMailEvery2h() {
	c := cron.New()
	_, err := c.AddFunc("@every 2h", func() {
		queryMailTask()
	})
	if err != nil {
		plog.Errorf("queryQQMailEvery2h error:%v", err)
	}
	c.Start()
	//go func() {
	//	queryMailTask()
	//}()
}

func queryMailTask() {
	println("处理任务..")
	var linkChan = make(chan string, 1)
	go func() {
		mail_fetch.FetchOssViolationUrl(linkChan)
	}()
	for s := range linkChan {
		println("fined msg:%v", s)
		success, _ := ossDeleteByFileUri(s)
		plog.Infof("删除文件:%v, 状态:%v", s, success)
	}
	println("结束任务..")
}

func ossDel(c *gin.Context) {
	param := struct {
		FileUri string `json:"uri" form:"uri"`
	}{}
	if err := c.ShouldBind(&param); err != nil || param.FileUri == "" {
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(400, gin.H{"error": "param empty"})
		}
		return
	}
	success, err := ossDeleteByFileUri(param.FileUri)
	if !success || err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"msg": "success"})
	plog.Infof("oss delete success Uri:%v", param.FileUri)

}

func ossDeleteByFileUri(fileUri string) (bool, error) {
	uri, err := url.Parse(fileUri)
	if err != nil {
		//c.JSON(400, gin.H{"uri parse error": err.Error()})
		return false, err
	}
	//mochatvid-9.oss-cn-beijing.aliyuncs.com
	hostName := uri.Host
	ossBucket := strings.Replace(hostName, "."+yourRegion, "", -1)
	if ossBucket == hostName {
		//c.JSON(400, gin.H{"uri param error": err.Error()})
		return false, errors.New("uri param error")
	}
	//snapshot/000E3FC6F4D4B4348D56369B831F01C3.jpg
	if uri.Path[0] != '/' {
		//c.JSON(400, gin.H{"path param error": err.Error()})
		return false, errors.New("path param error")
	}
	ossObjectKey := uri.Path[1:]
	plog.Infof("oss delete bucket:%v ossKey:%v", ossBucket, ossObjectKey)
	success, err := deleteObject(ossBucket, ossObjectKey)
	if !success || err != nil {
		//c.JSON(400, gin.H{"delete failed error": err.Error()})
		return false, errors.New(fmt.Sprintf("delete failed error:%v", err.Error()))
	}
	return true, nil
}

func deleteObject(yourBucketName, yourObjectKey string) (bool, error) {
	client, err := oss.New("oss-cn-beijing.aliyuncs.com", yourAccessKeyId, yourAccessKeySecret)

	if err != nil {
		fmt.Println("Error:", err)
		return false, err
	}

	// 获取存储空间。
	bucket, err := client.Bucket(yourBucketName)
	if err != nil {
		fmt.Println("Error:", err)
		return false, err
	}
	var exist bool
	var err1 error
	if exist, err1 = bucket.IsObjectExist(yourObjectKey); err1 != nil {
		fmt.Println("exist check Error:", err1)
		//return
	}
	fmt.Printf("文件(%v——%v): exist:%v\n", yourBucketName, yourObjectKey, exist)

	//删除文件。
	err = bucket.DeleteObject(yourObjectKey)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return false, err
	}
	fmt.Println("Object deleted")
	return true, nil
}
