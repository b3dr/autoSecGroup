package main

import (
	cloudsdk "acl/pkg/cloud-sdk"
	"acl/utils"
	"fmt"
	gin "github.com/gin-gonic/gin"
	ini "gopkg.in/ini.v1"
	"os"
)

const cfgFilePath = "conf.ini" // 配置文件地址
const DeviceName = "Auto-White-Policy"

func initClient() *cloudsdk.CloudSession {
	cfg, err := ini.Load(cfgFilePath) // 读配置文件
	if err != nil {
		fmt.Printf("Fail to read conf: %v", err)
		os.Exit(1)
	}

	client := new(cloudsdk.CloudSession)
	client.LoadCloudConf(cfg)

	return client
}

func main() {
	client := initClient()
	r := gin.Default()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"hello": "world"})
	})

	r.GET("/show", func(context *gin.Context) {
		context.String(200, client.Show())
	})

	r.GET("/change", func(context *gin.Context) {
		clientIP, err := utils.GetClientIPHelper(context.Request)
		if err != nil {
			panic(err)
		}

		ip := context.DefaultQuery("ip", clientIP)
		deviceName := context.DefaultQuery("name", DeviceName)

		context.String(200, client.ChangeRule(ip, deviceName))
	})

	r.GET("/ip", func(context *gin.Context) {
		ipv4, err := utils.GetClientIPHelper(context.Request)
		if err != nil {
			panic(err)
		}
		context.String(200, ipv4)
	})

	r.Run("127.0.0.1:6666")
}
