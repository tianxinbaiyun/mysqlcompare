package main

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/tianxinbaiyun/mysqlcompare/crontab"
	"github.com/tianxinbaiyun/mysqlcompare/service"
	"net/http"
)

func main() {

	// 定时任务
	go crontab.AddCron()

	// 添加路由
	engine := gin.New()
	engine.GET("/compare", func(context *gin.Context) {
		// 返回成功
		res := gin.H{
			"msg":    "success",
			"status": 0,
			"result": nil,
		}
		context.JSON(http.StatusOK, res)
	}) // 手动同步数据

	pprof.Register(engine)

	// 启用监听端口
	if err := engine.Run(":8080"); err != nil {
		panic(err)
	}

}

// Compare 对比
func Compare() {
	service.Compare()
}
