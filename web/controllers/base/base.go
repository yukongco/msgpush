package base

import (
	"github.com/gin-gonic/gin"
)

/* 发往前台的公共接口
*  statusCode : 状态码
*  Msg: 信息
 */
func WebResp(c *gin.Context, statusCode int, data interface{}, Msg string) {

	respMap := map[string]interface{}{"code": statusCode, "msg": Msg, "data": data}
	c.JSON(statusCode, respMap)

	return
}
