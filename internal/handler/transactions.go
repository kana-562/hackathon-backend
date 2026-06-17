package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/usecase"
)

type TransactionHandler struct {
	txUC usecase.TransactionUsecase
}

func NewTransactionHandler(txUC usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{txUC: txUC}
}

func (h *TransactionHandler) CreateTransaction(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req domain.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.StarterSetID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "starterSetId is required")
	}

	resp, err := h.txUC.CreateTransaction(userID, req.StarterSetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *TransactionHandler) GetTransaction(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	txID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid transaction id")
	}

	detail, err := h.txUC.GetTransaction(userID, txID)
	if err != nil {
		switch err.Error() {
		case "transaction not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized":
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *TransactionHandler) UpdateTransaction(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	txID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid transaction id")
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if body.Status == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	if err := h.txUC.UpdateStatus(userID, txID, body.Status); err != nil {
		switch err.Error() {
		case "transaction not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized":
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": body.Status})
}

func (h *TransactionHandler) GetStartPlan(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	txID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid transaction id")
	}

	plan, err := h.txUC.GetStartPlan(userID, txID)
	if err != nil {
		switch err.Error() {
		case "transaction not found", "set not found":
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case "unauthorized: only buyer can get start plan":
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, plan)
}
