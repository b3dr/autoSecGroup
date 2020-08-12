package cloud_sdk

import (
	"encoding/json"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	ini "gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
)

const PolicyName = "Auto-White-Policy"

var DefaultPort = "80,443"
var DefaultAction = "ACCEPT"
var DefaultProtocol = "TCP"

type TencentCloud struct {
	Client *vpc.Client
}

type response struct {
	SecurityGroupSet *secGroupResponse
}

type secGroupResponse struct {
}

type securityGroupSet struct {
	SecurityGroupId   string `json:"SecurityGroupId"`
	SecurityGroupName string `json:"SecurityGroupName"`
	SecurityGroupDesc string `json:"SecurityGroupDesc"`
	ProjectId         string `json:"ProjectId"`
	CreateTime        string `json:"CreateTime"`
}

func (tencent *TencentCloud) Register(cloudCfg *ini.File) {
	cfg := *cloudCfg.Section("Tencent")
	id := cfg.Key("secretId").String()
	key := cfg.Key("secretKey").String()
	credential := common.NewCredential(
		id,
		key,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(credential, "ap-hongkong", cpf)
	tencent.Client = client
}

func (tencent *TencentCloud) Show() string {
	response := tencent.querySecGroup()
	return response.ToJsonString()
}

func (tencent *TencentCloud) ChangeRule(targetIP string, deviceName string) string {
	message := "Target ip: " + targetIP + "\nDevice: " + deviceName + "\nResponse: "
	targetPort := DefaultPort

	fmt.Printf("\n当前设备IP： %v\n", targetIP)

	fmt.Printf("\n查询安全组...\n")
	secGroupResponse := tencent.querySecGroup()

	fmt.Printf("\n解析安全组查询结果...\n")
	secGroupId := tencent.parseGroupSetResponse(secGroupResponse)
	if len(secGroupId) == 0 {
		fmt.Printf("\n安全组不存在！请先设置安全组！\n")
		os.Exit(-2)
	}

	fmt.Printf("\n查询安全策略...\n")
	ingressPolicySet, err := tencent.queryPolicySet(secGroupId)
	if err != nil {
		fmt.Printf("\n查询安全策略失败！\n安全组ID： %v\n", secGroupId)
		panic(err)
	}

	fmt.Printf("\n解析安全策略...\n")
	isAutoPolicyExist := false // 标识位，判断自动策略是否在所有策略中
	for _, policy := range ingressPolicySet {
		isAutoPolicy, isTargetIP := tencent.parsePolicy(policy, targetIP, deviceName)

		// 当前策略是自动策略
		if isAutoPolicy {
			fmt.Printf("\n发现安全组自动配置策略！当前白名单IP： %v\n", *policy.CidrBlock)
			if !isTargetIP { // 存在自动策略，且目标ip不在策略中
				fmt.Printf("\n修改安全组策略...\n\n----------------------\n添加白名单IP：%v\n端口：%v\n协议：%v\n----------------------\n", targetIP, targetPort, DefaultProtocol)
				response, err := tencent.modifyPolicy(targetIP, targetPort, secGroupId, policy)
				if err == nil {
					fmt.Printf("\nResponse: \n%v\n", response)
					message += response
				} else {
					fmt.Printf("\nModify request error!\n")
					message += "Modify request error!"
				}
			} else { // 存在自动策略，且目标ip在策略中
				fmt.Printf("\nTarget policy is exist!\n")
				message += "Target policy is exist!"
			}
		}

		isAutoPolicyExist = isAutoPolicyExist || isAutoPolicy // 修改标记位
	}

	// TODO: 默认策略不存在，创建新策略
	if !isAutoPolicyExist {

	}

	return message
}

// 查询安全组
func (tencent *TencentCloud) querySecGroup() *vpc.DescribeSecurityGroupsResponse {

	request := vpc.NewDescribeSecurityGroupsRequest()

	params := "{}"
	err := request.FromJsonString(params)
	if err != nil {
		panic(err)
	}
	response, err := tencent.Client.DescribeSecurityGroups(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
	}
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", response.ToJsonString())
	return response
}

// 从安全组查询结果中解析安全组id
func (tencent *TencentCloud) parseGroupSetResponse(groupResponse *vpc.DescribeSecurityGroupsResponse) string {
	var securityGroupId string
	groupCount := *groupResponse.Response.TotalCount
	securityGroupSet := groupResponse.Response.SecurityGroupSet
	if groupCount >= 1 {
		securityGroup := securityGroupSet[0]
		securityGroupId = *securityGroup.SecurityGroupId
		fmt.Printf("\nFind security group rule: [ %v (%v) ] created at %v.\n",
			*securityGroup.SecurityGroupName, securityGroupId, *securityGroup.CreatedTime)
	} else {
		fmt.Printf("\nSecurity Group Set not exist!\n")
	}
	return securityGroupId
}

// 查询单个安全组的策略
func (tencent *TencentCloud) queryPolicySet(secGroupId string) ([]*vpc.SecurityGroupPolicy, error) {
	var ingressPolicySet []*vpc.SecurityGroupPolicy

	request := vpc.NewDescribeSecurityGroupPoliciesRequest()

	params := "{\"SecurityGroupId\":\"" + secGroupId + "\"}"
	err := request.FromJsonString(params)
	if err != nil {
		panic(err)
	}
	response, err := tencent.Client.DescribeSecurityGroupPolicies(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
	}
	if err == nil {
		ingressPolicySet = response.Response.SecurityGroupPolicySet.Ingress
	}

	return ingressPolicySet, err
}

// 解析单条策略
// 判断：1、是否是自动化策略；2、目标ip是否在规则中
func (tencent *TencentCloud) parsePolicy(policy *vpc.SecurityGroupPolicy, targetIP string, deviceName string) (bool, bool) {
	isAutoPolicy := false
	isTargetIP := false
	name := *policy.PolicyDescription
	if name == deviceName {
		isAutoPolicy = true
		cip := *policy.CidrBlock

		// 安全组策略中的ip不是所有ip，且目标ip在策略中
		if cip != "0.0.0.0/0" && isBelong(targetIP, cip) {
			isTargetIP = true
		}
	}

	return isAutoPolicy, isTargetIP
}

// TODO: 创建策略
func (tencent *TencentCloud) createPolicy(policy *vpc.SecurityGroupPolicy) string {
	policyParams, _ := json.Marshal(policy)

	return string(policyParams)
}

// 修改策略
func (tencent *TencentCloud) modifyPolicy(targetIP string, targetPort string, targetGroupId string, policy *vpc.SecurityGroupPolicy) (string, error) {
	*policy.CidrBlock = targetIP + "/24"
	*policy.Port = targetPort
	*policy.Protocol = DefaultProtocol
	*policy.Action = DefaultAction
	*policy.SecurityGroupId = targetGroupId // "sg-bekdu5br"

	policyParams, _ := json.Marshal(policy)
	policyString := string(policyParams)
	policyString = strings.ReplaceAll(policyString, "\"SecurityGroupId\":\""+*policy.SecurityGroupId+"\",", "")

	params := "{\"SecurityGroupId\":\"" + *policy.SecurityGroupId + "\",\"SecurityGroupPolicySet\":{\"Ingress\":[" + policyString + "]}}"
	params = strings.ReplaceAll(params, "\"ServiceTemplate\":{\"ServiceId\":\"\",\"ServiceGroupId\":\"\"},", "")
	params = strings.ReplaceAll(params, "\"AddressTemplate\":{\"AddressId\":\"\",\"AddressGroupId\":\"\"},", "")
	params = strings.ReplaceAll(params, "\"Ipv6CidrBlock\":\"\",", "")

	fmt.Printf("\nParams: \n%+v\n", params)

	request := vpc.NewReplaceSecurityGroupPolicyRequest()

	//params := "{\"SecurityGroupId\":\"sg-bekdu5br\",\"SecurityGroupPolicySet\":{\"Ingress\":[{\"PolicyIndex\":0,\"Protocol\":\"TCP\",\"Port\":\"80,443\",\"CidrBlock\":\"1.1.1.1/24\",\"SecurityGroupId\":\"sg-bekdu5br\",\"Action\":\"ACCEPT\",\"PolicyDescription\":\"Auto-White-Policy\"}]}}"
	err := request.FromJsonString(params)
	if err != nil {
		panic(err)
	}
	response, err := tencent.Client.ReplaceSecurityGroupPolicy(request)

	return response.ToJsonString(), err
}

func isBelong(ip, cidr string) bool {
	ipAddr := strings.Split(ip, `.`)
	if len(ipAddr) < 4 {
		return false
	}
	cidrArr := strings.Split(cidr, `/`)
	if len(cidrArr) < 2 {
		return false
	}
	var tmp = make([]string, 0)
	for key, value := range strings.Split(`255.255.255.0`, `.`) {
		iint, _ := strconv.Atoi(value)

		iint2, _ := strconv.Atoi(ipAddr[key])

		tmp = append(tmp, strconv.Itoa(iint&iint2))
	}
	return strings.Join(tmp, `.`) == cidrArr[0]
}
