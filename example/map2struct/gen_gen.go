// Auto generated code, DO NOT EDIT.

package map2struct

import (
	"github.com/adyzng/toolkit/example/model"
	"github.com/spf13/cast"
)

func genMapToBookInfo(src map[string]interface{}) (obj *model.ApiBookInfo, err error) {
	obj = &model.ApiBookInfo{}

	// 直接赋值的字段
	obj.Id = cast.ToInt64(src[""])
	obj.Name = cast.ToString(src[""])
	obj.CopyrightInfo = cast.ToString(src[""])
	obj.CreateTime = cast.ToString(src[""])
	obj.ThumbUrl = cast.ToString(src[""])
	obj.IsFirstRead = cast.ToBool(src[""])

	// 枚举类型
	if tmp, ok := src[""]; ok {
		val := (model.BookType)(cast.ToInt64(tmp))
		obj.BookType = &val
	}

	// 带赋值表达式的（指针类型）
	if tmp, ok := src[""]; ok {
		val := cast.ToInt32(tmp)
		obj.SerialCount = &val
	}
	if tmp, ok := src[""]; ok {
		val := cast.ToInt64(tmp)
		obj.LatestReadTime = &val
	}
	if tmp, ok := src[""]; ok {
		val := cast.ToString(tmp)
		obj.Category = &val
	}

	return obj, err
}
