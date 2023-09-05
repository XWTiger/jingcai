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
	CommentId uint
	Content   string
}

// 回复
type Response struct {
	gorm.Model
	//用户id （系统自动填）
	UserId    uint
	CommentId uint   `validate:"required"`
	Content   string `validate:"required"`
	//是否审核通过
	ShowStatus bool
}

func (res Response) getVO() ResponseVO {

	var vo = ResponseVO{
		Model:     res.Model,
		Content:   res.Content,
		CommentId: res.CommentId,
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
// @Router /bbs/comment [post]
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
	common.SuccessReturn(c, "评论成功")
}

// @Summary 回复评论
// @Description 回复评论
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param comment body Response  true  "回复内容"
// @Router /bbs/response [post]
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
// @Router /bbs/comment/list [get]
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
						respVo.UserVO = userRespMap[re.UserId]
						resVO = append(resVO, respVo)
					}
					commentVOs[i].Responses = resVO
				}
			}
			common.SuccessReturn(c, commentVOs)
			return
		}
	} else {
		common.FailedReturn(c, "参数错误")
		return
	}
}