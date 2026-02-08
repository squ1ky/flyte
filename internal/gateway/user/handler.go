package user

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/common"
	"github.com/squ1ky/flyte/pkg/api"
	"net/http"
)

type Handler struct {
	client userv1.UserServiceClient
}

func NewHandler(client userv1.UserServiceClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) SignUp(c *gin.Context) {
	var inp RegisterRequest
	if err := c.ShouldBindJSON(&inp); err != nil {
		api.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.Register(c.Request.Context(), &userv1.RegisterRequest{
		Email:       inp.Email,
		Password:    inp.Password,
		PhoneNumber: inp.PhoneNumber,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		UserID: resp.UserId,
	})
}

func (h *Handler) SignIn(c *gin.Context) {
	var inp LoginRequest
	if err := c.ShouldBindJSON(&inp); err != nil {
		api.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.Login(c.Request.Context(), &userv1.LoginRequest{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: resp.Token,
	})
}

func (h *Handler) AddPassenger(c *gin.Context) {
	targetUserID, ok := api.ParseID(c, "id")
	if !ok {
		return
	}

	if !h.checkOwnership(c, targetUserID) {
		return
	}

	var inp PassengerInput
	if err := c.ShouldBindJSON(&inp); err != nil {
		api.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.AddPassenger(c.Request.Context(), &userv1.AddPassengerRequest{
		UserId: targetUserID,
		Info: &userv1.Passenger{
			FirstName:      inp.FirstName,
			LastName:       inp.LastName,
			MiddleName:     inp.MiddleName,
			BirthDate:      inp.BirthDate,
			Gender:         inp.Gender,
			DocumentNumber: inp.DocumentNumber,
			DocumentType:   inp.DocumentType,
			Citizenship:    inp.Citizenship,
		},
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, AddPassengerResponse{
		PassengerID: resp.PassengerId,
	})
}

func (h *Handler) GetPassengers(c *gin.Context) {
	targetUserID, ok := api.ParseID(c, "id")
	if !ok {
		return
	}

	if !h.checkOwnership(c, targetUserID) {
		return
	}

	resp, err := h.client.GetPassengers(c.Request.Context(), &userv1.GetPassengersRequest{
		UserId: targetUserID,
	})
	if err != nil {
		api.HandleGRPCErr(c, err)
		return
	}

	passengers := make([]PassengerResponse, len(resp.Passengers))
	for i, p := range resp.Passengers {
		passengers[i] = mapProtoToPassenger(p)
	}

	c.JSON(http.StatusOK, passengers)
}

func (h *Handler) checkOwnership(c *gin.Context, requestedUserID int64) bool {
	tokenUserID, ok := common.GetUserID(c)
	if !ok {
		return false
	}

	if tokenUserID != requestedUserID {
		api.AbortForbidden(c)
		return false
	}
	return true
}

func mapProtoToPassenger(p *userv1.Passenger) PassengerResponse {
	return PassengerResponse{
		ID:             p.Id,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		MiddleName:     p.MiddleName,
		BirthDate:      p.BirthDate,
		Gender:         p.Gender,
		DocumentNumber: p.DocumentNumber,
		DocumentType:   p.DocumentType,
		Citizenship:    p.Citizenship,
	}
}
