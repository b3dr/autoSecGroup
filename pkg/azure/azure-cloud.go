package azure

import "gopkg.in/ini.v1"

type AzureCloud struct {
}

func (c *AzureCloud) Register(cloudCfg *ini.File) {

}

func (c *AzureCloud) Show() string {
	return ""
}

func (c *AzureCloud) ChangeRule(ip string, device string) string {
	return ""
}
