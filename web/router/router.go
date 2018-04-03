package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yukongco/msgpush/web/controllers/push"
)

var (
	RouteVersionOnePoint = "/push/v1"
)

func ApiRouter(router *gin.Engine) {

	version1 := router.Group(RouteVersionOnePoint)
	{
		admin := version1.Group("/admin")
		{
			admin.POST("/private", push.PushPrivate) // 系统单个推送
		}
	}
}
