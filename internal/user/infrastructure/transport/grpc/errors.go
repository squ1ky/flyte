package grpcserver

import (
	"errors"
	"github.com/squ1ky/flyte/internal/user/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MsgInvalidCredentials     = "invalid email or password"
	MsgUserNotFound           = "user not found"
	MsgUserAlreadyExists      = "user with this email already exists"
	MsgPassengerNotFound      = "passenger not found"
	MsgPassengerAlreadyExists = "passenger already exists"
)

func toGRPCError(err error) error {
	if domain.IsValidationError(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, MsgUserAlreadyExists)
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, MsgUserNotFound)
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, MsgInvalidCredentials)
	case errors.Is(err, domain.ErrPassengerNotFound):
		return status.Error(codes.NotFound, MsgPassengerNotFound)
	case errors.Is(err, domain.ErrPassengerAlreadyExists):
		return status.Error(codes.AlreadyExists, MsgPassengerAlreadyExists)
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
