package external

import (
	"context"
	"fmt"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/external/proto/token"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/model"
	"google.golang.org/grpc"
)

type Validation struct{}

func (u *Validation) ValidateToken(ctx context.Context, tokenReq string) (model.TokenResponse, error) {
	// connection to grpc
	conn, err := grpc.Dial("localhost:5000", grpc.WithInsecure())
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("failed to dial user service grpc :%w", err)
	}
	defer conn.Close()

	client := token.NewTokenServiceClient(conn)

	req := token.TokenRequest{
		Token: tokenReq,
	}

	response, err := client.Validate(ctx, &req)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("failed to validate token : %w", err)
	}

	if response.Message != "Token Valid" {
		return model.TokenResponse{}, fmt.Errorf("got response error from user service grpc :%s", response.Message)
	}

	return model.TokenResponse{
		UserID: response.Data.Id,
		Email:  response.Data.Email,
	}, nil
}
