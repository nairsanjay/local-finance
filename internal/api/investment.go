package api

import (
	"github.com/gin-gonic/gin"
	"local-finance/internal/service"
	"net/http"
)

func (h *Handler) ListInvestments(c *gin.Context) {
	result, err := h.db.ListInvestmentSnapshots()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load investments"})
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *Handler) ImportInvestment(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxInvestmentFileSize+(1<<20))
	if err := c.Request.ParseMultipartForm(service.MaxInvestmentFileSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload or file exceeds 10 MB"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "choose an investment statement file"})
		return
	}
	defer file.Close()
	snapshot, duplicate, err := service.NewInvestmentService(h.db).Import(header.Filename, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"snapshot": snapshot, "duplicate": duplicate})
}
func (h *Handler) DeleteInvestment(c *gin.Context) {
	deleted, err := h.db.DeleteInvestmentSnapshot(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete investment snapshot"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "investment snapshot not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
