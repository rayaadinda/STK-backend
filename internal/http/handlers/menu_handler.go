package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"stk-backend/internal/menu"
	"stk-backend/internal/response"
)

type MenuHandler struct {
	service *menu.Service
}

func NewMenuHandler(service *menu.Service) *MenuHandler {
	return &MenuHandler{service: service}
}

type createMenuRequest struct {
	Name     string  `json:"name" binding:"required,max=120"`
	ParentID *string `json:"parentId"`
}

type updateMenuRequest struct {
	Name string `json:"name" binding:"required,max=120"`
}

type moveMenuRequest struct {
	ParentID *string `json:"parentId"`
}

type reorderMenuRequest struct {
	ParentID *string `json:"parentId"`
	Position *int    `json:"position" binding:"required"`
}

func (h *MenuHandler) GetAll(c *gin.Context) {
	tree, err := h.service.GetTree(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.JSONSuccess(c, http.StatusOK, tree, "menus retrieved")
}

func (h *MenuHandler) GetByID(c *gin.Context) {
	menuID, ok := parsePathMenuID(c)
	if !ok {
		return
	}

	node, err := h.service.GetByID(c.Request.Context(), menuID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.JSONSuccess(c, http.StatusOK, node, "menu retrieved")
}

func (h *MenuHandler) Create(c *gin.Context) {
	var request createMenuRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.JSONError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request payload", gin.H{"details": err.Error()})
		return
	}

	if request.ParentID != nil {
		trimmedParentID := strings.TrimSpace(*request.ParentID)
		if trimmedParentID == "" {
			request.ParentID = nil
		} else {
			request.ParentID = &trimmedParentID
			if err := menu.ValidateUUID(trimmedParentID); err != nil {
				response.JSONError(c, http.StatusBadRequest, "INVALID_PARENT_ID", "parentId must be a valid UUID", nil)
				return
			}
		}
	}

	created, err := h.service.Create(c.Request.Context(), menu.CreateInput{
		Name:     request.Name,
		ParentID: request.ParentID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.JSONSuccess(c, http.StatusCreated, created, "menu created")
}

func (h *MenuHandler) Update(c *gin.Context) {
	menuID, ok := parsePathMenuID(c)
	if !ok {
		return
	}

	var request updateMenuRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.JSONError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request payload", gin.H{"details": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), menuID, menu.UpdateInput{
		Name: request.Name,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.JSONSuccess(c, http.StatusOK, updated, "menu updated")
}

func (h *MenuHandler) Delete(c *gin.Context) {
	menuID, ok := parsePathMenuID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), menuID); err != nil {
		h.handleError(c, err)
		return
	}

	response.JSONSuccess(c, http.StatusOK, gin.H{"id": menuID}, "menu deleted")
}

func (h *MenuHandler) Move(c *gin.Context) {
	menuID, ok := parsePathMenuID(c)
	if !ok {
		return
	}

	var request moveMenuRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.JSONError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request payload", gin.H{"details": err.Error()})
		return
	}

	if request.ParentID != nil {
		trimmedParentID := strings.TrimSpace(*request.ParentID)
		if trimmedParentID == "" {
			request.ParentID = nil
		} else {
			request.ParentID = &trimmedParentID
			if err := menu.ValidateUUID(trimmedParentID); err != nil {
				response.JSONError(c, http.StatusBadRequest, "INVALID_PARENT_ID", "parentId must be a valid UUID", nil)
				return
			}
		}
	}

	moved, err := h.service.Move(c.Request.Context(), menuID, menu.MoveInput{
		ParentID: request.ParentID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.JSONSuccess(c, http.StatusOK, moved, "menu moved")
}

func (h *MenuHandler) Reorder(c *gin.Context) {
	menuID, ok := parsePathMenuID(c)
	if !ok {
		return
	}

	var request reorderMenuRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.JSONError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request payload", gin.H{"details": err.Error()})
		return
	}

	if request.Position == nil {
		response.JSONError(c, http.StatusBadRequest, "INVALID_REQUEST", "position is required", nil)
		return
	}

	if request.ParentID != nil {
		trimmedParentID := strings.TrimSpace(*request.ParentID)
		if trimmedParentID == "" {
			request.ParentID = nil
		} else {
			request.ParentID = &trimmedParentID
			if err := menu.ValidateUUID(trimmedParentID); err != nil {
				response.JSONError(c, http.StatusBadRequest, "INVALID_PARENT_ID", "parentId must be a valid UUID", nil)
				return
			}
		}
	}

	reordered, err := h.service.Reorder(c.Request.Context(), menuID, menu.ReorderInput{
		ParentID: request.ParentID,
		Position: *request.Position,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.JSONSuccess(c, http.StatusOK, reordered, "menu reordered")
}

func parsePathMenuID(c *gin.Context) (string, bool) {
	menuID := strings.TrimSpace(c.Param("id"))
	if menuID == "" {
		response.JSONError(c, http.StatusBadRequest, "INVALID_MENU_ID", "menu id is required", nil)
		return "", false
	}

	if err := menu.ValidateUUID(menuID); err != nil {
		response.JSONError(c, http.StatusBadRequest, "INVALID_MENU_ID", "menu id must be a valid UUID", nil)
		return "", false
	}

	return menuID, true
}

func (h *MenuHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, menu.ErrMenuNotFound):
		response.JSONError(c, http.StatusNotFound, "MENU_NOT_FOUND", "menu not found", nil)
	case errors.Is(err, menu.ErrParentMenuNotFound):
		response.JSONError(c, http.StatusBadRequest, "PARENT_NOT_FOUND", "parent menu not found", nil)
	case errors.Is(err, menu.ErrMenuNameInvalid):
		response.JSONError(c, http.StatusBadRequest, "INVALID_MENU_NAME", "menu name cannot be empty", nil)
	case errors.Is(err, menu.ErrInvalidMoveTarget):
		response.JSONError(c, http.StatusBadRequest, "INVALID_MOVE_TARGET", "cannot move menu to itself or its descendants", nil)
	case errors.Is(err, menu.ErrInvalidPosition):
		response.JSONError(c, http.StatusBadRequest, "INVALID_POSITION", "position must be greater than or equal to zero", nil)
	default:
		response.JSONError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error", gin.H{"details": err.Error()})
	}
}
