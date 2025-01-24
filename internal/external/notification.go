package external

import (
	"context"
	"fmt"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/env"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/external/proto/notification"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type NotifRequest struct {
	Recipient    string
	TemplateName string
	Placeholder  map[string]string
}

type notif struct {
}

func (n *notif) SendNotification(ctx context.Context, req NotifRequest) error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	// connection to grpc
	conn, err := grpc.Dial(env.GetEnvString("NOTIF_SERVICE", ""), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to dial user service grpc :%w", err)
	}
	defer conn.Close()

	client := notification.NewNotificationServiceClient(conn)

	request := &notification.SendNotificationRequest{
		Recipient:    req.Recipient,
		TemplateName: req.TemplateName,
		Placeholder:  req.Placeholder,
	}

	resp, err := client.SendNotification(ctx, request)
	if err != nil {
		return err
	}

	if resp.Message != "success" {
		return fmt.Errorf("got response error :%s", resp.Message)
	}

	return nil
}
