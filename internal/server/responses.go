package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success sends a successful JSON response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	RespondWithError(c, http.StatusConflict, ErrorTypeInvalidRequest, message)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	RespondWithError(c, http.StatusServiceUnavailable, ErrorTypeServerError, message)
}