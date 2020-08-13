package tencent

import (
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"strings"
)

// 处理腾讯云API请求参数
// SDK默认生成的请求参数中包含这些字段，但这些字段会产生冲突，即使为空也不行，需要把字段删掉
func GenerateRequestParams(params string) string {
	params = strings.ReplaceAll(params, "\"ServiceTemplate\":{\"ServiceId\":\"\",\"ServiceGroupId\":\"\"},", "")
	params = strings.ReplaceAll(params, "\"AddressTemplate\":{\"AddressId\":\"\",\"AddressGroupId\":\"\"},", "")
	params = strings.ReplaceAll(params, "\"Ipv6CidrBlock\":\"\",", "")
	return params
}

// 处理腾讯云API响应错误
func ParseResponseErr(err error) {
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
	}
	if err != nil {
		panic(err)
	}
}
