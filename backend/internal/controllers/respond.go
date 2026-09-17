package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
)

func respond(c *gin.Context, err error) { apierr.Respond(c, err) }
