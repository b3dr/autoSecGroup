package vultr

import "gopkg.in/ini.v1"

type VultrCloud struct {
}

func (c *VultrCloud) Register(cloudCfg *ini.File) {

}

func (c *VultrCloud) Show() string {
	return ""
}

func (c *VultrCloud) ChangeRule(ip string, device string) string {
	return ""
}
