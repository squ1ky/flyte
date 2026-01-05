package handler

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"net/http"
)

type UserHandler struct {
	client userv1.UserServiceClient
}

func NewUserHandler(client userv1.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

type registerInput struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phone_number"`
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var inp registerInput
	if err := c.BindJSON(&inp); err != nil {
		newErrorResponse(c, http.StatusBadRequest, ErrInvalidInputBody)
		return
	}

	resp, err := h.client.Register(c.Request.Context(), &userv1.RegisterRequest{
		Email:       inp.Email,
		Password:    inp.Password,
		PhoneNumber: inp.PhoneNumber,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": resp.UserId})
}

type loginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) SignIn(c *gin.Context) {
	var inp loginInput
	if err := c.BindJSON(&inp); err != nil {
		newErrorResponse(c, http.StatusBadRequest, ErrInvalidInputBody)
		return
	}

	resp, err := h.client.Login(c.Request.Context(), &userv1.LoginRequest{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": resp.Token})
}

type passengerInput struct {
	FirstName      string `json:"first_name" binding:"required"`
	LastName       string `json:"last_name" binding:"required"`
	MiddleName     string `json:"middle_name"`
	BirthDate      string `json:"birth_date" binding:"required"`
	Gender         string `json:"gender" binding:"required,oneof=male female"`
	DocumentNumber string `json:"document_number" binding:"required"`
	DocumentType   string `json:"document_type" binding:"required"`
	Citizenship    string `json:"citizenship" binding:"required,len=3"`
}

func (h *UserHandler) AddPassenger(c *gin.Context) {
	targetUserID, err := parseIDParam(c, "id")
	if err != nil {
		return
	}

	if !verifyUserOwnership(c, targetUserID) {
		return
	}

	var inp passengerInput
	if err := c.BindJSON(&inp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
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
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"passenger_id": resp.PassengerId})
}

func (h *UserHandler) GetPassengers(c *gin.Context) {
	targetUserID, err := parseIDParam(c, "id")
	if err != nil {
		return
	}

	if !verifyUserOwnership(c, targetUserID) {
		return
	}

	resp, err := h.client.GetPassengers(c.Request.Context(), &userv1.GetPassengersRequest{
		UserId: targetUserID,
	})
	if err != nil {
		mapGRPCErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp.Passengers)
}
