package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"hobby-relay-backend/internal/usecase"
)

type MypageHandler struct {
	mypageUC usecase.MypageUsecase
}

func NewMypageHandler(mypageUC usecase.MypageUsecase) *MypageHandler {
	return &MypageHandler{mypageUC: mypageUC}
}

func (h *MypageHandler) GetMe(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	user, err := h.mypageUC.GetMe(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (h *MypageHandler) GetMyPage(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.mypageUC.GetMyPage(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *MypageHandler) GetSelling(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	cards, err := h.mypageUC.GetSelling(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"sets": cards})
}

func (h *MypageHandler) GetPurchases(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	purchases, err := h.mypageUC.GetPurchases(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"transactions": purchases})
}

func (h *MypageHandler) GetFavorites(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	favorites, err := h.mypageUC.GetFavorites(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"sets": favorites})
}
