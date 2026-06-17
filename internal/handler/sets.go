package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/usecase"
)

type SetsHandler struct {
	setsUC usecase.SetsUsecase
}

func NewSetsHandler(setsUC usecase.SetsUsecase) *SetsHandler {
	return &SetsHandler{setsUC: setsUC}
}

func (h *SetsHandler) GetSets(c echo.Context) error {
	userID := getUserID(c)

	q := c.QueryParam("q")
	smart := c.QueryParam("smart") == "true"
	categoryID, _ := strconv.ParseInt(c.QueryParam("categoryId"), 10, 64)
	hobbyID, _ := strconv.ParseInt(c.QueryParam("hobbyId"), 10, 64)
	sort := c.QueryParam("sort")
	maxPrice, _ := strconv.Atoi(c.QueryParam("maxPrice"))
	minBeginner, _ := strconv.Atoi(c.QueryParam("minBeginnerScore"))
	minReadiness, _ := strconv.Atoi(c.QueryParam("minReadinessScore"))

	query := domain.SearchQuery{
		Q:                 q,
		Smart:             smart,
		CategoryID:        categoryID,
		HobbyID:           hobbyID,
		Sort:              sort,
		MaxPrice:          maxPrice,
		MinBeginnerScore:  minBeginner,
		MinReadinessScore: minReadiness,
	}

	resp, err := h.setsUC.Search(query, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *SetsHandler) GetCategorySets(c echo.Context) error {
	userID := getUserID(c)
	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid category id")
	}

	cards, err := h.setsUC.GetCategorySets(categoryID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"sets": cards})
}

func (h *SetsHandler) GetSetDetail(c echo.Context) error {
	userID := getUserID(c)
	setID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid set id")
	}

	detail, err := h.setsUC.GetSetDetail(setID, userID)
	if err != nil {
		if err.Error() == "set not found" {
			return echo.NewHTTPError(http.StatusNotFound, "set not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *SetsHandler) AddFavorite(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}
	setID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid set id")
	}

	if err := h.setsUC.AddFavorite(userID, setID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]bool{"favorited": true})
}

func (h *SetsHandler) RemoveFavorite(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}
	setID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid set id")
	}

	if err := h.setsUC.RemoveFavorite(userID, setID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]bool{"favorited": false})
}

func (h *SetsHandler) AskQuestion(c echo.Context) error {
	setID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid set id")
	}

	var req domain.SetQuestionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Message == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message is required")
	}

	answer, err := h.setsUC.AskQuestion(setID, req.Message)
	if err != nil {
		if err.Error() == "set not found" {
			return echo.NewHTTPError(http.StatusNotFound, "set not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, domain.SetQuestionResponse{Message: answer})
}
