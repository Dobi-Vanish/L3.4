package handler

import (
	"io"
	"net/http"

	"L3.4/internal/service"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

type Handler struct {
	imageService *service.ImageService
	log          logger.Logger
}

func NewHandler(imageService *service.ImageService, log logger.Logger) *Handler {
	return &Handler{
		imageService: imageService,
		log:          log,
	}
}

func (h *Handler) RegisterRoutes(r *ginext.Engine) {
	api := r.Group("/api")
	{
		api.POST("/upload", h.uploadHandler)
		api.GET("/images", h.listHandler)
		api.GET("/image/:id", h.getImageHandler)
		api.DELETE("/image/:id", h.deleteHandler)
	}

	r.Static("/web", "./web")
	r.GET("/", func(c *ginext.Context) {
		c.File("./web/index.html")
	})
}

func (h *Handler) uploadHandler(c *ginext.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "no file uploaded"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to read file"})
		return
	}

	img, err := h.imageService.UploadImage(c.Request.Context(), fileHeader.Filename, data)
	if err != nil {
		h.log.Error("upload failed", "error", err)
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	h.log.Info("image uploaded", "id", img.ID, "status", img.Status)
	c.JSON(http.StatusOK, ginext.H{
		"id":     img.ID,
		"status": img.Status,
	})
}

func (h *Handler) listHandler(c *ginext.Context) {
	images, err := h.imageService.ListImages(c.Request.Context())
	if err != nil {
		h.log.Error("list failed", "error", err)
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	h.log.Info("list images", "count", len(images))
	c.JSON(http.StatusOK, images)
}

func (h *Handler) getImageHandler(c *ginext.Context) {
	id := c.Param("id")
	variant := c.DefaultQuery("variant", "original")

	img, err := h.imageService.GetImage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	if img == nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "image not found"})
		return
	}

	path, err := h.imageService.GetImageFile(img, variant)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}
	if path == "" {
		c.JSON(http.StatusNotFound, ginext.H{"error": "variant not available"})
		return
	}

	c.File(path)
}

func (h *Handler) deleteHandler(c *ginext.Context) {
	id := c.Param("id")
	if err := h.imageService.DeleteImage(c.Request.Context(), id); err != nil {
		h.log.Error("delete failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"status": "deleted"})
}
