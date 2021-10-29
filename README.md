
## 使用方法

自动生成 map 到结构的转换代码，支持的 map value 类型有：基本类型 + interface{}

待生成函数签名要求：入参必须有 `map[string]xxx` 类型（只用第一个），返回值第一个必须是待转换的结构体

### 1. 下载 map2struct

``` bash
go get github.com/adyzng/toolkit/cmd/map2struct
go install github.com/adyzng/toolkit/cmd/map2struct
```

### 2. 在工程目录下配置 go:generate

``` go
// 示例文件: map2struct.go

package gen

import (
    "context"

	"github.com/adyzng/xxx/common_base"
)

//go:generate map2struct
// 会枚举此文件中的所有`MapTo`开头的 function, 在 _gen.go 文件中生成对应的转换函数
// 函数签名要求：入参必须有`map[string]xxx`类型（只用第一个），返回值第一个必须是待转换的结构体（可以有多个返回值）
// 直接命令行也行(需要在当前工程目录下): map2struct -input="./gen/map2struct.go"

func MapToApiItemInfo(src map[string]string) *common_base.ApiItemInfo {
	return nil
}

// map[string]interface{} 也可以支持
func MapToApiItemInfo2(src map[string]interface{}) (*common_base.ApiItemInfo, error) {
	return nil, nil
}

func MapToApiItemInfo3(src map[string]string, param interface{}) (*common_base.ApiItemInfo, error) {
	return nil, nil
}

func MapToApiBookInfo(ctx context.Context, isWeb bool, src map[string]string) (*common_base.ApiBookInfo, error) {
	return  nil, nil
}
```

### 3. 运行 go generate

``` go
// 生成代码如下：map2struct_gen.go
// Auto generated code, DO NOT EDIT.

package gen

import "github.com/adyzng/xxx/common_base"

func genMapToApiItemInfo(src map[string]string) (obj *common_base.ApiItemInfo, err error) {
	obj = &common_base.ApiItemInfo{}
    ...
	return obj, err
}

func genMapToApiBookInfo(src map[string]string) (obj *common_base.ApiBookInfo, err error) {
	obj = &common_base.ApiBookInfo{}
    ...
	return obj, err
}
```