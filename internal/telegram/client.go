package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"

	"github.com/FolcloreX/GopherGram/internal/config"
)

type Client struct {
	client   *telegram.Client
	sender   *message.Sender
	uploader *uploader.Uploader

	phone    string
	password string
	appID    int
	appHash  string
	chatID   int64

	TargetPeer tg.InputPeerClass
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		phone:    cfg.Phone,
		password: cfg.Password,
		appID:    cfg.APIID,
		appHash:  cfg.APIHash,
		chatID:   cfg.ChatID,
	}
}

func (c *Client) Start(ctx context.Context, runLogic func(ctx context.Context) error) error {
	// Configure persistent session (session.json)
	sessionDir := "session"
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return err
	}
	sessionStorage := &session.FileStorage{
		Path: filepath.Join(sessionDir, "session.json"),
	}

	// Waits the time specified in the FloodAwait
	waiter := floodwait.NewSimpleWaiter()

	c.client = telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: sessionStorage,
		Middlewares: []telegram.Middleware{
			waiter,
		},
	})

	return c.client.Run(ctx, func(ctx context.Context) error {
		if err := c.authenticate(ctx); err != nil {
			return err
		}

		raw := c.client.API()
		c.sender = message.NewSender(raw)
		c.uploader = uploader.NewUploader(raw)

		fmt.Println("✅ Userbot conectado e autenticado!")

		return runLogic(ctx)
	})
}

// Authenticate manages the login (Phone -> Code -> 2FA Password)
func (c *Client) authenticate(ctx context.Context) error {
	flow := auth.NewFlow(
		auth.Constant(c.phone, c.password, auth.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
			fmt.Printf("📲 Digite o código enviado para %s: ", c.phone)
			var code string
			fmt.Scan(&code)
			return code, nil
		})),
		auth.SendCodeOptions{},
	)

	return c.client.Auth().IfNecessary(ctx, flow)
}
