package common

// BaseResponse 返回对象
type BaseResponse struct {
	//1 成功 0 失败
	Code int `json:"code"`

	//错误信息
	Message string `json:"message"`

	// in: body
	Content interface{} `json:"content"`
}

// PageCL 分页
type PageCL struct {
	//页码
	PageNo int
	//每页大小
	PageSize int
	//总条数
	Total int
	//内容
	Content interface{}
}

func Success(c interface{}) *BaseResponse {
	return &BaseResponse{
		Code:    1,
		Message: "执行成功",
		Content: c,
	}
}

func Failed() *BaseResponse {
	return &BaseResponse{
		Code:    0,
		Message: "执行失败",
	}
}
