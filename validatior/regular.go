package validatior

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"jingcai/common"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	REX  = "regular"
	MSG  = "message"
	MUST = "validate"
	MIN  = "min"
	MAX  = "max"
)

func Validator(c *gin.Context, dest any) {

	s := reflect.TypeOf(dest)

	v := reflect.ValueOf(dest)
	for i := 0; i < s.NumField(); i++ {
		fmt.Println(s.Field(i).Tag)
		tag := s.Field(i).Tag
		t := s.Field(i).Type
		//必填校验
		if tag.Get(MUST) != "" {
			switch t.Kind() {
			case reflect.Int:
				if v.Field(i).Int() <= 0 {
					common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
					return
				}
				if tag.Get(MIN) != "" {
					min, err := strconv.ParseInt(tag.Get(MIN), 10, 64)
					if err == nil {
						if v.Field(i).Int() < min {
							common.FailedReturn(c, fmt.Sprintf("%s 必须大于%v", s.Field(i).Name, min))
							return
						}
					}
				}

				if tag.Get(MAX) != "" {
					max, err := strconv.ParseInt(tag.Get(MAX), 10, 64)
					if err == nil {
						if v.Field(i).Int() > max {
							common.FailedReturn(c, fmt.Sprintf("%s 必须小于%v", s.Field(i).Name, max))
							return
						}
					}
				}

				break
			case reflect.String:
				if v.Field(i).Len() <= 0 {
					common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
					return
				}
				break
			case reflect.Float32, reflect.Float64:
				if v.Field(i).Float() <= 0 {
					common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
					return
				}
				break
			}

		}

		//正则校验
		if tag.Get(REX) != "" {
			match, err := regexp.MatchString(tag.Get(REX), v.Field(i).String())
			if err != nil || match == false {
				if tag.Get(MSG) != "" {
					common.FailedReturn(c, tag.Get(MSG))
					return
				} else {
					common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 校验失败"))
					return
				}
			}
		}

		fmt.Println("name: ", t.Name(), "kind: ", t.Kind(), "str: ", t.String())
		if strings.Compare(t.Kind().String(), "struct") == 0 {
			Validator(c, v.Field(i).Interface())
		}

		if strings.Compare(t.Kind().String(), "slice") == 0 {
			sValue := v.Field(i)
			for j := 0; j < sValue.Len(); j++ {
				Validator(c, sValue.Index(j).Interface())
			}
		}

	}
}
