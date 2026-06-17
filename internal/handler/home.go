package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"hobby-relay-backend/internal/usecase"
)

type HomeHandler struct {
	homeUC usecase.HomeUsecase
}

func NewHomeHandler(homeUC usecase.HomeUsecase) *HomeHandler {
	return &HomeHandler{homeUC: homeUC}
}

func (h *HomeHandler) GetHome(c echo.Context) error {
	userID := getUserID(c)
	resp, err := h.homeUC.GetHome(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *HomeHandler) GetCategories(c echo.Context) error {
	categories, err := h.homeUC.GetCategories()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, categories)
}

// getUserID extracts optional user ID from context (0 if not authenticated)
func getUserID(c echo.Context) int64 {
	val := c.Get("userId")
	if val == nil {
		return 0
	}
	id, ok := val.(int64)
	if !ok {
		return 0
	}
	return id
}

// requireUserID extracts required user ID or returns error
func requireUserID(c echo.Context) (int64, error) {
	val := c.Get("userId")
	if val == nil {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	id, ok := val.(int64)
	if !ok || id == 0 {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	return id, nil
}
