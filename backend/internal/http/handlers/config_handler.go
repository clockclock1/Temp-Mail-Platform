package handlers

import (
	"net/http"

	"tempmail/backend/internal/config"
	"tempmail/backend/internal/runtime"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	cfgManager *config.Manager
	runtimeCtl *runtime.Controller
}

func NewConfigHandler(cfgManager *config.Manager, runtimeCtl *runtime.Controller) *ConfigHandler {
	return &ConfigHandler{cfgManager: cfgManager, runtimeCtl: runtimeCtl}
}

func (h *ConfigHandler) Get(c *gin.Context) {
	cfg := h.cfgManager.Get()
	c.JSON(http.StatusOK, gin.H{"item": cfg})
}

func (h *ConfigHandler) Update(c *gin.Context) {
	oldCfg := h.cfgManager.Get()
	newCfg := oldCfg
	if err := c.ShouldBindJSON(&newCfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.cfgManager.Update(newCfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appliedCfg := h.cfgManager.Get()
	result, err := h.runtimeCtl.Apply(oldCfg, appliedCfg)
	if err != nil {
		_ = h.cfgManager.Update(oldCfg)
		_, _ = h.runtimeCtl.Apply(appliedCfg, oldCfg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item":            appliedCfg,
		"warnings":        result.Warnings,
		"restartRequired": result.RestartRequired,
	})
}

func (h *ConfigHandler) Reload(c *gin.Context) {
	oldCfg := h.cfgManager.Get()
	if err := h.cfgManager.Reload(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newCfg := h.cfgManager.Get()
	result, err := h.runtimeCtl.Apply(oldCfg, newCfg)
	if err != nil {
		_ = h.cfgManager.Update(oldCfg)
		_, _ = h.runtimeCtl.Apply(newCfg, oldCfg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"item":            newCfg,
		"warnings":        result.Warnings,
		"restartRequired": result.RestartRequired,
	})
}
