package cloud_sdk

import (
	"acl/pkg/tencent"
	"acl/pkg/vultr"
	"fmt"
	ini "gopkg.in/ini.v1"
	"os"
)

// 云厂商的SDK接口
type Cloud interface {
	// 利用云厂商key注册Session
	Register(cfg *ini.File)

	// 利用Session调用API获取基本信息
	Show() string

	// 利用Session调用API修改ACL Rule
	ChangeRule(ip string, device string) string
}

type CloudSession struct {
	TencentClient tencent.TencentCloud
	VultrClient   vultr.VultrCloud
	CloudType     string
	Cloud
}

func (c *CloudSession) LoadCloudConf(cfg *ini.File) {
	// todo: ini文件在这里解析，解析出Section，作为副本或者指针传到函数里，发现Section里只有key没有value？
	// todo：只能把整个File的副本或指针传到函数里在进行解析

	// 判断云厂商类型
	c.CloudType = cfg.Section("cloud").Key("cloud").In("Tencent", []string{"Tencent", "Vultr"})
	switch cloudType := c.CloudType; {

	// 腾讯云client
	case cloudType == "Tencent":
		c.TencentClient.Register(cfg)
		c.Cloud = &c.TencentClient // 实例化之后的云厂商SDK Session传递给统一的SDK接口

	// Vultr client
	case cloudType == "Vultr":
		c.VultrClient.Register(cfg)
		c.Cloud = &c.VultrClient

	default:
		fmt.Printf("Fail to find conf for: %v", c.CloudType)
		os.Exit(1)
	}
}
