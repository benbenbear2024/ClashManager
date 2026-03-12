package api

import (
	"clash-manager/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	nodeHandler := handlers.NewNodeHandler()
	ruleHandler := handlers.NewRuleHandler()
	sourceHandler := handlers.NewSubscriptionSourceHandler()

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
	}
}
