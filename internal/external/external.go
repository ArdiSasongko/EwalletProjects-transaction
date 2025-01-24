package external

import (
	"context"
	"net/http"
	"time"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/model"
)

type External struct {
	Notif interface {
		SendNotification(context.Context, NotifRequest) error
	}
	Wallet interface {
		Credit(context.Context, WalletRequest, string) (*WalletResponse, error)
		Debit(context.Context, WalletRequest, string) (*WalletResponse, error)
	}
	Validation interface {
		ValidateToken(context.Context, string) (model.TokenResponse, error)
	}
}

func NewExternal() External {
	return External{
		Notif: &notif{},
		Wallet: &wallet{
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		},
		Validation: &Validation{},
	}
}
