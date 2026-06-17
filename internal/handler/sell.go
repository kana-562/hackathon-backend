package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/usecase"
)

type SellHandler struct {
	sellUC usecase.SellUsecase
}

func NewSellHandler(sellUC usecase.SellUsecase) *SellHandler {
	return &SellHandler{sellUC: sellUC}
}

func (h *SellHandler) CreateDraft(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req domain.CreateDraftRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	resp, err := h.sellUC.CreateDraft(userID, req.HobbyText)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *SellHandler) SendMessage(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	draftID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid draft id")
	}

	var req domain.SellMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Message == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message is required")
	}

	resp, err := h.sellUC.SendMessage(userID, draftID, req.Message)
	if err != nil {
		if err.Error() == "unauthorized" {
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *SellHandler) GetDraft(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	draftID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid draft id")
	}

	detail, err := h.sellUC.GetDraft(userID, draftID)
	if err != nil {
		switch err.Error() {
		case "draft not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized":
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *SellHandler) UpdateDraft(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	draftID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid draft id")
	}

	var req domain.UpdateDraftRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.sellUC.UpdateDraft(userID, draftID, req); err != nil {
		switch err.Error() {
		case "draft not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized":
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]bool{"updated": true})
}

func (h *SellHandler) PublishDraft(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	draftID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid draft id")
	}

	if err := h.sellUC.PublishDraft(userID, draftID); err != nil {
		switch err.Error() {
		case "draft not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized":
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "on_sale"})
}
