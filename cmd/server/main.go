package main

import (
	"clash-manager/internal/api"
	"clash-manager/internal/utils"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"clash-manager/web"

	"github.com/gin-gonic/gin"
)

func main() {
	portPtr := flag.String("port", "8090", "Server port (default: 8090)")
	flag.Parse()

	serverPort := formatPort(*portPtr)

	// 检查端口是否被占用，如果是则强制结束进程
	portNum, err := strconv.Atoi(strings.TrimPrefix(serverPort, ":"))
	if err == nil {
		killed, err := utils.CheckAndKillPortProcess(portNum)
		if err != nil {
			log.Printf("Warning: %v", err)
		} else if killed {
			log.Printf("Successfully killed process occupying port %d", portNum)
		}
	}

	r := gin.Default()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	api.SetupRoutes(r)

	subFS, _ := fs.Sub(web.StaticFiles, "dist")
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/sub/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		f, err := subFS.Open(path[1:])
		if err != nil {
			indexFile, err := subFS.Open("index.html")
			if err != nil {
				c.String(http.StatusNotFound, "index.html not found")
				return
			}
			defer indexFile.Close()

			stat, _ := indexFile.Stat()
			c.DataFromReader(http.StatusOK, stat.Size(), "text/html; charset=utf-8", indexFile, nil)
			return
		}
		defer f.Close()

		stat, _ := f.Stat()
		contentType := getContentType(path)
		c.DataFromReader(http.StatusOK, stat.Size(), contentType, f, nil)
	})

	log.Printf("Server starting on %s...", serverPort)
	if err := r.Run(serverPort); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func formatPort(port string) string {
	if port == "" {
		return ":8090"
	}
	if port[0] == ':' {
		return port
	}
	return ":" + port
}

func getContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(path, ".gif"):
		return "image/gif"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(path, ".woff"):
		return "font/woff"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(path, ".ttf"):
		return "font/ttf"
	case strings.HasSuffix(path, ".eot"):
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}
