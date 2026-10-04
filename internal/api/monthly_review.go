package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"local-finance/internal/db"
)

func (h *Handler) GetMonthlyReview(c *gin.Context) {
	review, err := h.db.GetMonthlyReview(c.Query("month"), time.Now())
	if err != nil {
		if errors.Is(err, db.ErrInvalidReviewPeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load monthly review"})
		}
		return
	}
	c.JSON(http.StatusOK, review)
}

func (h *Handler) GetMonthlyReviewEvidence(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	period := c.DefaultQuery("period", "current")
	if err != nil || page < 1 || page > 1000000 || (period != "current" && period != "previous") ||
		!c.Request.URL.Query().Has("category") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provide a category, a valid page, and period current or previous"})
		return
	}
	evidence, err := h.db.GetMonthlyReviewEvidence(c.Query("month"), c.Query("category"), period == "previous", page, time.Now())
	if err != nil {
		if errors.Is(err, db.ErrInvalidReviewPeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load supporting transactions"})
		}
		return
	}
	c.JSON(http.StatusOK, evidence)
}
