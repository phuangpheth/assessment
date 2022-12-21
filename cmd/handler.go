package cmd

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/phuangpheth/assessment/track"
)

type handler struct {
	expenseSvc *track.Service
}

func NewHandler(router *echo.Echo, svc *track.Service) error {
	if router == nil || svc == nil {
		return errors.New("invalid argument")
	}
	h := handler{
		expenseSvc: svc,
	}

	router.POST("/expenses", h.SaveExpense)
	return nil
}

func (h *handler) SaveExpense(c echo.Context) error {
	var exp track.Expense
	if err := c.Bind(&exp); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"code":    http.StatusBadRequest,
			"message": "invalid request body",
		})
	}
	if err := exp.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
	}

	ctx := c.Request().Context()
	expense, err := h.expenseSvc.Save(ctx, &exp)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"code":    http.StatusInternalServerError,
			"message": "Internal Server Error: ",
		})
	}
	return c.JSON(http.StatusCreated, expense)
}
