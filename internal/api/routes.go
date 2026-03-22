package api

import (
	"clash-manager/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	nodeHandler := handlers.NewNodeHandler()
	ruleHandler := handlers.NewRuleHandler()
	sourceHandler := handlers.NewSubscriptionSourceHandler()
	settingsHandler := handlers.NewSettingsHandler()
	mihomoHandler := handlers.NewMihomoHandler()

	api := r.Group("/api")
	{
		api.GET("/nodes", nodeHandler.ListNodes)
		api.POST("/nodes", nodeHandler.CreateNode)
		api.POST("/nodes/import", nodeHandler.ImportNode)
		api.PUT("/nodes/:id", nodeHandler.UpdateNode)
		api.DELETE("/nodes/:id", nodeHandler.DeleteNode)
		api.GET("/nodes/:id/export", nodeHandler.ExportNode)

		api.GET("/rules", ruleHandler.ListRules)
		api.GET("/rules/tags", ruleHandler.GetTags)
		api.POST("/rules", ruleHandler.CreateRule)
		api.POST("/rules/import", ruleHandler.ImportRules)
		api.PUT("/rules/:id", ruleHandler.UpdateRule)
		api.DELETE("/rules/:id", ruleHandler.DeleteRule)

		api.GET("/sources", sourceHandler.ListSources)
		api.GET("/sources/:id", sourceHandler.GetSource)
		api.POST("/sources", sourceHandler.CreateSource)
		api.PUT("/sources/:id", sourceHandler.UpdateSource)
		api.DELETE("/sources/:id", sourceHandler.DeleteSource)
		api.POST("/sources/:id/sync", sourceHandler.SyncSource)
		api.POST("/sources/test", sourceHandler.TestSource)

		api.GET("/settings/dns", settingsHandler.GetDNS)
		api.POST("/settings/dns", settingsHandler.SaveDNS)

		// Mihomo 控制
		api.GET("/settings/mihomo/status", mihomoHandler.GetMihomoStatus)
		api.POST("/settings/mihomo/start", mihomoHandler.StartMihomo)
		api.POST("/settings/mihomo/stop", mihomoHandler.StopMihomo)
		api.POST("/settings/mihomo/reload", mihomoHandler.ReloadMihomo)
		api.POST("/settings/mihomo/check-connection", mihomoHandler.CheckL2TPConnection)

		// 配置文件编辑
		api.GET("/settings/config/content", mihomoHandler.GetConfigContent)
		api.POST("/settings/config/content", mihomoHandler.SaveConfigContent)

		// 系统信息
		api.GET("/system/info", handlers.SystemHandler())
	}
}
