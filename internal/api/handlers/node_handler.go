package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"clash-manager/internal/model"
	"clash-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	Service *service.NodeService
}

func NewNodeHandler() *NodeHandler {
	return &NodeHandler{Service: service.NewNodeService()}
}

func (h *NodeHandler) ListNodes(c *gin.Context) {
	nodes, err := h.Service.ListNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

func (h *NodeHandler) CreateNode(c *gin.Context) {
	var node model.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.CreateNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, node)
}

func (h *NodeHandler) UpdateNode(c *gin.Context) {
	var node model.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		fmt.Printf("UpdateNode: JSON binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idStr := c.Param("id")
	fmt.Printf("UpdateNode: Received ID parameter: '%s'\n", idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Printf("UpdateNode: ID conversion error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	node.ID = uint(id)

	fmt.Printf("UpdateNode: Updating node ID=%d, Name=%s, Type=%s, Server=%s, Port=%d, Username=%s, Password=%s, Rename=%s\n", 
		node.ID, node.Name, node.Type, node.Server, node.Port, node.Username, node.Password, node.Rename)

	if err := h.Service.UpdateNode(uint(id), &node); err != nil {
		fmt.Printf("UpdateNode: Update error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("UpdateNode: Successfully updated node ID=%d\n", node.ID)
	c.JSON(http.StatusOK, node)
}

func (h *NodeHandler) DeleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.Service.DeleteNode(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *NodeHandler) ImportNode(c *gin.Context) {
	var req struct {
		Link string `json:"link"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	node, err := service.ParseLink(req.Link)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link format: " + err.Error()})
		return
	}

	existingNodes, err := h.Service.ListNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get existing nodes"})
		return
	}

	existingNames := make(map[string]bool)
	for _, n := range existingNodes {
		existingNames[n.Name] = true
	}

	originalName := node.Name
	newName := ""
	for j := 1; ; j++ {
		newName = fmt.Sprintf("Name%d", j)
		if !existingNames[newName] {
			break
		}
	}

	node.Rename = originalName
	node.Name = newName

	if err := h.Service.CreateNode(node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save node: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, node)
}

func (h *NodeHandler) ExportNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	node, err := h.Service.GetNodeByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	link, err := service.ExportLink(node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export node: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"link": link,
		"name": node.Name,
		"type": node.Type,
	})
}
