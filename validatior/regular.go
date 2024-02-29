package validatior

import (
	"errors"
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

// `validate:"required" message:"需要期号" regular: "正则"`
func Validator(c *gin.Context, dest any) error {

	s := reflect.TypeOf(dest)

	v := reflect.ValueOf(dest)
	for i := 0; i < s.NumField(); i++ {
		fmt.Println(s.Field(i).Tag)
		tag := s.Field(i).Tag
		t := s.Field(i).Type
		fmt.Println("name: ", t.Name(), "kind: ", t.Kind(), "str: ", t.String())
		//必填校验
		if tag.Get(MUST) != "" {
			switch t.Kind() {
			case reflect.Int:
				fmt.Println(" value: ", v.Field(i).Int())
				if v.Field(i).Int() <= 0 {
					if c != nil {
						common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
					}
					c.Abort()
					return errors.New(fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
				}
				if tag.Get(MIN) != "" {
					min, err := strconv.ParseInt(tag.Get(MIN), 10, 64)
					if err == nil {
						if v.Field(i).Int() < min {
							if c != nil {
								common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
							}
							c.Abort()
							return errors.New(fmt.Sprintf("%s 必须大于%v", s.Field(i).Name, min))
						}
					}
				}

				if tag.Get(MAX) != "" {
					max, err := strconv.ParseInt(tag.Get(MAX), 10, 64)
					if err == nil {
						if v.Field(i).Int() > max {
							if c != nil {
								common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
							}
							c.Abort()
							return errors.New(fmt.Sprintf("%s 必须小于%v", s.Field(i).Name, max))
						}
					}
				}

				break
			case reflect.String:
				fmt.Println(" value: ", v.Field(i).String())
				if v.Field(i).Len() <= 0 || v.Field(i).String() == "" {
					if c != nil {
						common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
					}
					c.Abort()
					return errors.New(fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
				}
				break
			case reflect.Float32, reflect.Float64:
				fmt.Println(" value: ", v.Field(i).Float())
				if v.Field(i).Float() <= 0 {
					if c != nil {
						common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
					}
					c.Abort()
					return errors.New(fmt.Sprintf("%s %s", s.Field(i).Name, " 必填"))
				}
				break
			case reflect.Bool:
				fmt.Println(" value: ", v.Field(i).Bool())
			}

		}

		//正则校验
		if tag.Get(REX) != "" {
			match, err := regexp.MatchString(tag.Get(REX), v.Field(i).String())
			if err != nil || match == false {
				if tag.Get(MSG) != "" {
					if c != nil {
						common.FailedReturn(c, tag.Get(MSG))
					}
					c.Abort()
					return errors.New(tag.Get(MSG))
				} else {
					if c != nil {
						common.FailedReturn(c, fmt.Sprintf("%s %s", s.Field(i).Name, " 校验失败"))
					}
					c.Abort()
					return errors.New(fmt.Sprintf("%s %s", s.Field(i).Name, " 校验失败"))
				}
			}
		}

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
	return nil
}
