package bbs

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"jingcai/validatior"
	"strconv"
)

// 评论
type CommentVO struct {
	gorm.Model
	user.UserVO
	//帖子id
	ContentID uint
	Content   string
	Responses []ResponseVO
}

type ResponseVO struct {
	gorm.Model
	user.UserVO
	CommentId    uint
	Content      string
	ResponseId   uint
	ResponseName string
}

// 回复
type Response struct {
	gorm.Model
	//用户id （系统自动填）
	UserId     uint
	ResponseId uint
	CommentId  uint   `validate:"required"`
	Content    string `validate:"required"`
	//是否审核通过
	ShowStatus bool
}

func (res Response) getVO() ResponseVO {

	var vo = ResponseVO{
		Model:      res.Model,
		Content:    res.Content,
		CommentId:  res.CommentId,
		ResponseId: res.ResponseId,
	}
	return vo
}

func (com Comment) getCommentVO() CommentVO {
	var vo = CommentVO{
		Model:     com.Model,
		ContentID: com.ContentID,
		Content:   com.Content,
	}
	return vo
}

// 评论
type Comment struct {
	gorm.Model
	//用户id （系统自动填）
	UserId uint
	//帖子id
	ContentID uint `validate:"required"`
	//评论内容
	Content string `validate:"required"`
	//是否审核通过
	ShowStatus bool
}

// @Summary 评论帖子
// @Description 评论帖子
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param comment body Comment  true  "评论内容"
// @Router /api/bbs/comment [post]
func CommentHandler(c *gin.Context) {
	var comment Comment
	err := c.BindJSON(&comment)
	if err != nil {
		common.FailedReturn(c, "获取参数失败")
		return
	}
	validatior.Validator(c, comment)
	var userInfo = user.FetUserInfo(c)
	comment.UserId = userInfo.ID
	comment.ShowStatus = true
	if err := mysql.DB.Model(Comment{}).Save(&comment).Error; err != nil {
		log.Error(err)
		common.FailedReturn(c, "评论失败")
		return
	}
	common.SuccessReturn(c, comment.ID)
}

// @Summary 回复评论
// @Description 回复评论
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param comment body Response  true  "回复内容"
// @Router /api/bbs/response [post]
func ResponseHandler(c *gin.Context) {
	var response Response
	err := c.BindJSON(&response)
	if err != nil {
		common.FailedReturn(c, "获取参数失败")
		return
	}
	var userInfo = user.FetUserInfo(c)
	response.ShowStatus = true
	response.UserId = userInfo.ID
	validatior.Validator(c, response)
	if err := mysql.DB.Model(Response{}).Save(&response).Error; err != nil {
		log.Error(err)
		common.FailedReturn(c, "评论失败")
		return
	}
	common.SuccessReturn(c, "评论成功")
}

// @Summary  查询回复和评论
// @Description 通过帖子id 查询回复和评论
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @param bbsId query int  true  "每页条数"
// @Router /api/bbs/comment/list [get]
func ListComment(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")
	id := c.Query("bbsId")
	var comments []Comment
	if pageNo != "" && pageSize != "" && id != "" {
		pageN, _ := strconv.Atoi(pageNo)
		pageS, _ := strconv.Atoi(pageSize)
		contentId, _ := strconv.Atoi(id)
		var count int64
		mysql.DB.Model(Comment{}).Where(&Comment{ContentID: uint(contentId)}).Count(&count)
		mysql.DB.Model(Comment{}).Where(&Comment{ContentID: uint(contentId)}).Offset((pageN - 1) * pageS).Limit(pageS).Find(&comments)
		if len(comments) == 0 {
			common.SuccessReturn(c, comments)
			return
		} else {
			var comIds []uint
			for _, comment := range comments {
				comIds = append(comIds, comment.UserId)
			}
			userMap := user.FindUsserMapById(comIds)
			var commentVOs []CommentVO
			for _, comment := range comments {
				vo := comment.getCommentVO()
				vo.UserVO = userMap[comment.UserId]
				commentVOs = append(commentVOs, vo)
			}
			for i := 0; i < len(commentVOs); i++ {
				var res []Response
				mysql.DB.Model(Response{}).Where(&Response{CommentId: commentVOs[i].ID}).Find(&res)
				if len(res) > 0 {
					var resVO []ResponseVO
					var resPonseUseIds []uint
					for _, re := range res {
						resPonseUseIds = append(resPonseUseIds, re.UserId)
					}
					userRespMap := user.FindUsserMapById(resPonseUseIds)
					for _, re := range res {
						respVo := re.getVO()
						if respVo.ResponseId > 0 {
							result := getResponseById(respVo.ResponseId, res)
							if result != nil {
								respVo.ResponseName = userRespMap[result.UserId].Name
							}

						}

						respVo.UserVO = userRespMap[re.UserId]
						resVO = append(resVO, respVo)
					}
					commentVOs[i].Responses = resVO
				}
			}

			common.SuccessReturn(c, common.PageCL{
				PageNo:   pageN,
				PageSize: pageS,
				Total:    int(count),
				Content:  commentVOs,
			})
			return
		}
	} else {
		common.FailedReturn(c, "参数错误")
		return
	}

}

// @Summary  通过评论id查回复
// @Description 通过帖子id 查询回复和评论
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param commentId query int  true  "评论id"
// @Router /api/bbs/comment/response [get]
func GetResponseByCommentId(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")

	id := c.Query("commentId")
	commentId, err := strconv.Atoi(id)
	if err != nil {
		log.Error(err)
		common.FailedReturn(c, "参数异常")
		return
	}
	var comment Comment
	if err := mysql.DB.Model(Comment{}).Where(&Comment{Model: gorm.Model{ID: uint(commentId)}}).First(&comment).Error; err != nil {
		common.FailedReturn(c, "评论不存在")
		return
	}
	var res []Response
	mysql.DB.Model(Response{}).Where(&Response{CommentId: uint(commentId)}).Find(&res)

	var resVO []ResponseVO
	var resPonseUseIds []uint
	for _, re := range res {
		resPonseUseIds = append(resPonseUseIds, re.UserId)
	}
	resPonseUseIds = append(resPonseUseIds, comment.UserId)
	userRespMap := user.FindUsserMapById(resPonseUseIds)
	for _, re := range res {
		respVo := re.getVO()
		if respVo.ResponseId > 0 {
			result := getResponseById(respVo.ResponseId, res)
			if result != nil {
				respVo.ResponseName = userRespMap[result.UserId].Name
			}

		}

		respVo.UserVO = userRespMap[re.UserId]
		resVO = append(resVO, respVo)
	}
	comVO := comment.getCommentVO()
	comUser := userRespMap[comment.UserId]
	comVO.Name = comUser.Name
	comVO.Phone = comUser.Phone
	comVO.Avatar = comUser.Avatar
	comVO.Responses = resVO
	common.SuccessReturn(c, comVO)

}

func getResponseById(id uint, list []Response) *Response {
	for _, response := range list {
		if response.ID == id {
			return &response
		}
	}
	return nil
}
