package routes

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/config"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/controllers"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/middleware"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

type Options struct {
	Config    *config.Config
	DB        *gorm.DB
	Validator *jwtx.Validator
	Deps      controllers.Deps
	MediaDir  string
	Log       *slog.Logger
}

func Register(r *gin.Engine, o Options) {
	r.Use(middleware.Recovery(o.Log), middleware.Logger(o.Log), middleware.CORS(o.Config.CORSOrigins))

	// Locally generated media is served straight off disk. In production this
	// would be object storage; the storage interface already allows the swap.
	r.Static("/media", o.MediaDir)

	r.GET("/health", func(c *gin.Context) {
		sql, err := o.DB.DB()
		if err != nil || sql.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "database": "unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "env": o.Config.Env})
	})

	models := controllers.NewModels(o.Deps)
	gens := controllers.NewGenerations(o.Deps)
	keys := controllers.NewKeys(o.Deps)
	me := controllers.NewMe(o.Deps)

	v1 := r.Group("/api/v1")
	{
		// Public: the landing page's model wall needs the catalogue before login.
		v1.GET("/catalog", models.Public)

		auth := v1.Group("")
		auth.Use(middleware.Auth(o.Validator, o.DB))
		{
			auth.GET("/me", me.Get)
			auth.GET("/models", models.List)

			auth.POST("/generations", gens.Create)
			auth.GET("/generations", gens.List)
			auth.GET("/generations/:id", gens.Get)
			auth.POST("/generations/:id/cancel", gens.Cancel)

			auth.GET("/keys", keys.List)
			auth.PUT("/keys", keys.Save)
			auth.DELETE("/keys/:provider", keys.Delete)
		}
	}
}
