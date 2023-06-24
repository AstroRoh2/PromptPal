package routes

import (
	"net/http"
	"time"

	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	ErrorCode    int    `json:"code"`
	ErrorMessage string `json:"error"`
}

func SetupGinRoutes(commitSha string) *gin.Engine {
	h := gin.Default()

	// with version
	h.Use(func(c *gin.Context) {
		c.Writer.Header().Add("X-PP-VER", commitSha)
		c.Next()
	})
	h.Use(brotli.Brotli(brotli.DefaultCompression))

	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://*.annatarhe.com", "http://*.annatarhe.cn"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "X-Token", "Content-Type", "Authorization", "Content-Encoding", "Date", "X-RSA-Auth", "X-RSA-Nonce"},
		ExposeHeaders:    []string{"Content-Length", "Content-Encoding", "Date"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	h.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: "page not found",
		})
	})

	authRoutes := h.Group("/api/v1/auth")
	authRoutes.POST("/login", authHandler)

	adminRoutes := h.Group("/api/v1/admin")
	adminRoutes.Use(authMiddleware)
	{

		adminRoutes.GET("/users", listUsers)
		adminRoutes.POST("/users", createUsers)
		adminRoutes.DELETE("/users/:id", removeUsers)

		adminRoutes.GET("/projects", listProjects)
		adminRoutes.GET("/projects/:id", getProject)
		adminRoutes.POST("/projects", createProject)
		adminRoutes.PUT("/projects/:id", updateProject)

		adminRoutes.GET("/prompts", listPrompts)
		adminRoutes.GET("/prompts/:id", getPrompt)
		adminRoutes.POST("/prompts", createPrompt)
		adminRoutes.PUT("/prompts/:id", updatePrompt)
	}

	apiRoutes := h.Group("/api/v1/public")
	apiRoutes.Use(apiMiddleware)
	{
		apiRoutes.GET("/prompts", apiListPrompts)
		apiRoutes.POST("/prompts/run/:id", apiRunPrompt)
	}

	return h
}
