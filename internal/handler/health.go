package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// CheckHealth godoc
// @Summary Check service health
// @Description Verify API and database availability.
// @Tags Health
// @Produce json
// @Success 200 {object} response.BaseResponse{data=response.HealthResponse}
// @Failure 500 {object} response.ErrorResponseDTO
// @Router /api/health [get]
func (hh *HealthHandler) CheckHealth(c *gin.Context) {
	sql, err := hh.db.DB()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "db_error"})
	}

	if err := sql.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "db_down"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
