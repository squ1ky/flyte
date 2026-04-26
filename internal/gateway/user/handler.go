package user

import (
	"github.com/gin-gonic/gin"
	userv1 "github.com/squ1ky/flyte/gen/proto/user"
	"github.com/squ1ky/flyte/internal/gateway/common"
	"net/http"
	"time"
)

const (
	paramUserID      = "id"
	paramPassengerID = "passengerId"
)

type Handler struct {
	client userv1.UserServiceClient
}

func NewHandler(client userv1.UserServiceClient) *Handler {
	return &Handler{client: client}
}

// SignUp godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new user account with the provided email, password, and phone number.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		RegisterRequest		true	"Registration data"
//	@Success		201		{object}	RegisterResponse	"User created successfully"
//	@Failure		400		{object}	common.ErrorResponse		"Invalid input"
//	@Failure		409		{object}	common.ErrorResponse		"Email already taken"
//	@Failure		500		{object}	common.ErrorResponse		"Internal server error"
//	@Router			/auth/sign-up [post]
func (h *Handler) SignUp(c *gin.Context) {
	var inp RegisterRequest
	if err := c.ShouldBindJSON(&inp); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.Register(c.Request.Context(), &userv1.RegisterRequest{
		Email:       inp.Email,
		Password:    inp.Password,
		PhoneNumber: inp.PhoneNumber,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		UserID: resp.UserId,
	})
}

// SignIn godoc
//
//	@Summary		Authenticate user
//	@Description	Authenticates a user by email and password, returns a JWT token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		LoginRequest	true	"Login credentials"
//	@Success		200		{object}	LoginResponse	"Authenticated successfully"
//	@Failure		400		{object}	common.ErrorResponse	"Invalid input"
//	@Failure		401		{object}	common.ErrorResponse	"Invalid credentials"
//	@Failure		500		{object}	common.ErrorResponse	"Internal server error"
//	@Router			/auth/sign-in [post]
func (h *Handler) SignIn(c *gin.Context) {
	var inp LoginRequest
	if err := c.ShouldBindJSON(&inp); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.Login(c.Request.Context(), &userv1.LoginRequest{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		UserID: resp.UserId,
		Token:  resp.Token,
	})
}

// GetUser godoc
//
//	@Summary		Get user profile
//	@Description	Returns profile information for the specified user. Only the account owner can access this.
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int				true	"User ID"
//	@Success		200	{object}	UserResponse	"User profile"
//	@Failure		403	{object}	common.ErrorResponse	"Forbidden — not the account owner"
//	@Failure		404	{object}	common.ErrorResponse	"User not found"
//	@Failure		500	{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	targetUserID, ok := common.ParseID(c, paramUserID)
	if !ok {
		return
	}

	if !h.checkOwnership(c, targetUserID) {
		return
	}

	resp, err := h.client.GetUser(c.Request.Context(), &userv1.GetUserRequest{
		UserId: targetUserID,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	var createdAt string
	if resp.CreatedAt != nil {
		createdAt = resp.CreatedAt.AsTime().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:          resp.Id,
		Email:       resp.Email,
		PhoneNumber: resp.PhoneNumber,
		Role:        common.RoleFromProto(resp.Role),
		CreatedAt:   createdAt,
	})
}

// AddPassenger godoc
//
//	@Summary		Add a passenger
//	@Description	Creates a new passenger record linked to the specified user.
//	@Tags			passengers
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"User ID"
//	@Param			input	body		PassengerInput			true	"Passenger data"
//	@Success		201		{object}	AddPassengerResponse	"Passenger created"
//	@Failure		400		{object}	common.ErrorResponse			"Invalid input"
//	@Failure		403		{object}	common.ErrorResponse			"Forbidden — not the account owner"
//	@Failure		409		{object}	common.ErrorResponse			"Duplicate passenger document"
//	@Failure		500		{object}	common.ErrorResponse			"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id}/passengers [post]
func (h *Handler) AddPassenger(c *gin.Context) {
	targetUserID, ok := common.ParseID(c, paramUserID)
	if !ok {
		return
	}

	if !h.checkOwnership(c, targetUserID) {
		return
	}

	var inp PassengerInput
	if err := c.ShouldBindJSON(&inp); err != nil {
		common.AbortInvalidInput(c, err)
		return
	}

	resp, err := h.client.AddPassenger(c.Request.Context(), &userv1.AddPassengerRequest{
		UserId: targetUserID,
		Info: &userv1.PassengerInfo{
			FirstName:      inp.FirstName,
			LastName:       inp.LastName,
			MiddleName:     inp.MiddleName,
			BirthDate:      inp.BirthDate,
			Gender:         genderToProto(inp.Gender),
			DocumentNumber: inp.DocumentNumber,
			DocumentType:   inp.DocumentType,
			Citizenship:    inp.Citizenship,
		},
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, AddPassengerResponse{
		PassengerID: resp.PassengerId,
	})
}

// GetPassengers godoc
//
//	@Summary		List passengers
//	@Description	Returns all passengers belonging to the specified user.
//	@Tags			passengers
//	@Produce		json
//	@Param			id	path		int					true	"User ID"
//	@Success		200	{array}		PassengerResponse	"List of passengers"
//	@Failure		403	{object}	common.ErrorResponse		"Forbidden — not the account owner"
//	@Failure		500	{object}	common.ErrorResponse		"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id}/passengers [get]
func (h *Handler) GetPassengers(c *gin.Context) {
	targetUserID, ok := common.ParseID(c, paramUserID)
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
		common.HandleGRPCError(c, err)
		return
	}

	passengers := make([]PassengerResponse, len(resp.Passengers))
	for i, p := range resp.Passengers {
		passengers[i] = mapProtoToPassenger(p)
	}

	c.JSON(http.StatusOK, passengers)
}

// DeletePassenger godoc
//
//	@Summary		Delete a passenger
//	@Description	Removes a passenger record belonging to the specified user.
//	@Tags			passengers
//	@Param			id			path	int	true	"User ID"
//	@Param			passengerId	path	int	true	"Passenger ID"
//	@Success		204			"Passenger deleted"
//	@Failure		403			{object}	common.ErrorResponse	"Forbidden — not the account owner"
//	@Failure		404			{object}	common.ErrorResponse	"Passenger not found"
//	@Failure		500			{object}	common.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id}/passengers/{passengerId} [delete]
func (h *Handler) DeletePassenger(c *gin.Context) {
	targetUserID, ok := common.ParseID(c, paramUserID)
	if !ok {
		return
	}

	if !h.checkOwnership(c, targetUserID) {
		return
	}

	passengerID, ok := common.ParseID(c, paramPassengerID)
	if !ok {
		return
	}

	_, err := h.client.DeletePassenger(c.Request.Context(), &userv1.DeletePassengerRequest{
		UserId:      targetUserID,
		PassengerId: passengerID,
	})
	if err != nil {
		common.HandleGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// checkOwnership verifies that the token user ID matches the requested resource owner.
// Returns false and aborts the request with 403 Forbidden if they don't match.
func (h *Handler) checkOwnership(c *gin.Context, requestedUserID int64) bool {
	tokenUserID, ok := common.GetUserID(c)
	if !ok {
		return false
	}

	if tokenUserID != requestedUserID {
		common.AbortForbidden(c)
		return false
	}
	return true
}
