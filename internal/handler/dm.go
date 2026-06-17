package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/usecase"
)

type DMHandler struct {
	dmUC usecase.DMUsecase
}

func NewDMHandler(dmUC usecase.DMUsecase) *DMHandler {
	return &DMHandler{dmUC: dmUC}
}

func (h *DMHandler) GetOrCreateRoom(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req domain.GetOrCreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.PartnerID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "partnerId is required")
	}

	resp, err := h.dmUC.GetOrCreateRoom(userID, req.PartnerID, req.SetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *DMHandler) ListRooms(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	rooms, err := h.dmUC.ListRooms(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, rooms)
}

func (h *DMHandler) GetMessages(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
	}

	msgs, err := h.dmUC.GetMessages(userID, roomID)
	if err != nil {
		switch err.Error() {
		case "room not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized":
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, msgs)
}

func (h *DMHandler) SendMessage(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
	}

	var req domain.SendDMRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	msg, err := h.dmUC.SendMessage(userID, roomID, req.Body)
	if err != nil {
		switch err.Error() {
		case "room not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized":
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, msg)
}

func (h *DMHandler) MarkRead(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
	}

	if err := h.dmUC.MarkRead(userID, roomID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}
