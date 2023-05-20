// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/bbs/commit": {
            "post": {
                "description": "提交论坛的帖子",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "提交帖子",
                "parameters": [
                    {
                        "description": "提交对象",
                        "name": "param",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/bbs.BBS"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    }
                }
            }
        },
        "/bbs/list": {
            "get": {
                "description": "提交论坛的帖子",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "查询全部贴子",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "日期 unix time",
                        "name": "date",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "页码",
                        "name": "pageNo",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "每页条数",
                        "name": "pageSize",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "状态检测",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/salt": {
            "get": {
                "description": "公钥 默认10分钟过期",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "获取加密的公钥",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    }
                }
            }
        },
        "/super/creep": {
            "get": {
                "description": "爬虫接口",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "爬虫接口",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user": {
            "post": {
                "description": "创建用户",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "创建用户",
                "parameters": [
                    {
                        "description": "用户对象",
                        "name": "param",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.UserVO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    }
                }
            }
        },
        "/user/login": {
            "post": {
                "description": "公钥放在头里 salt， 密码：需要和公钥rsa 加密 账号为手机号",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "登录接口",
                "parameters": [
                    {
                        "description": "用户对象",
                        "name": "param",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.UserVO"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/common.BaseResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "bbs.BBS": {
            "type": "object",
            "properties": {
                "bbsContent": {
                    "description": "论坛信息",
                    "$ref": "#/definitions/creeper.Content"
                },
                "userInfo": {
                    "description": "用户信息",
                    "$ref": "#/definitions/user.User"
                }
            }
        },
        "common.BaseResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "1 成功 0 失败",
                    "type": "integer"
                },
                "content": {
                    "description": "in: body"
                },
                "message": {
                    "description": "错误信息",
                    "type": "string"
                }
            }
        },
        "creeper.Content": {
            "type": "object",
            "properties": {
                "conditions": {
                    "description": "条件 让球 1.25",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "content": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "type": "string"
                },
                "extra": {
                    "description": "额外的一些信息",
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "imageUrl": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "match": {
                    "description": "比赛",
                    "type": "string"
                },
                "predict": {
                    "description": "预测谁赢",
                    "type": "string"
                },
                "summery": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                },
                "type": {
                    "description": "网站名称",
                    "type": "string",
                    "example": "雷速"
                },
                "updatedAt": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                },
                "userID": {
                    "type": "integer"
                }
            }
        },
        "user.User": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "description": "昵称",
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "role": {
                    "description": "\"enum: Admin,User\"",
                    "type": "string"
                },
                "salt": {
                    "description": "盐",
                    "type": "string"
                },
                "secret": {
                    "description": "密码",
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "user.UserVO": {
            "type": "object",
            "properties": {
                "name": {
                    "description": "昵称",
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "secret": {
                    "description": "密码",
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
