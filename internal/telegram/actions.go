package telegram

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

func (c *Client) UploadAndSendVideo(ctx context.Context, filePath string, caption string) error {
	fileName := filepath.Base(filePath)
	fmt.Printf("⬆️  [Vídeo] Iniciando upload: %s\n", fileName)

	fileUpload, err := c.uploader.FromPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("falha upload video '%s': %w", fileName, err)
	}

	p := &tg.InputPeerChannel{ChannelID: c.chatID, AccessHash: 0}

	_, err = c.sender.To(p).Video(ctx, fileUpload, html.String(nil, caption))

	if err != nil {
		return fmt.Errorf("erro envio video '%s': %w", fileName, err)
	}

	fmt.Printf("✅ [Vídeo] Enviado: %s\n", fileName)
	time.Sleep(2 * time.Second)
	return nil
}

func (c *Client) UploadAndSendDocument(ctx context.Context, filePath string, caption string) error {
	fileName := filepath.Base(filePath)
	fmt.Printf("⬆️  [Arquivo] Iniciando upload: %s\n", fileName)

	fileUpload, err := c.uploader.FromPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("falha upload arquivo '%s': %w", fileName, err)
	}

	p := &tg.InputPeerChannel{ChannelID: c.chatID, AccessHash: 0}

	_, err = c.sender.To(p).File(ctx, fileUpload, html.String(nil, caption))

	if err != nil {
		return fmt.Errorf("erro envio arquivo '%s': %w", fileName, err)
	}

	fmt.Printf("✅ [Arquivo] Enviado: %s\n", fileName)
	time.Sleep(2 * time.Second)
	return nil
}

// SendMessage passes the index using HTML
func (c *Client) SendMessage(ctx context.Context, text string) (int, error) {
	updates, err := c.sender.To(&tg.InputPeerChannel{
		ChannelID:  c.chatID,
		AccessHash: 0,
	}).StyledText(ctx, html.String(nil, text))

	if err != nil {
		return 0, fmt.Errorf("falha envio index: %w", err)
	}

	return extractMsgID(updates), nil
}

// PinMessage uses the RAW API
func (c *Client) PinMessage(ctx context.Context, messageID int) error {
	_, err := c.client.API().MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:   &tg.InputPeerChannel{ChannelID: c.chatID, AccessHash: 0},
		ID:     messageID,
		Silent: true,
	})

	return err
}

func (c *Client) CheckChatAccess(ctx context.Context) error {
	_, err := c.client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: c.chatID, AccessHash: 0},
	})
	if err != nil {
		fmt.Printf("⚠️ Sincronizando diálogos...\n")
		_, err := c.client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{Limit: 100})
		if err != nil {
			return err
		}
	}
	return nil
}

func extractMsgID(updates tg.UpdatesClass) int {
	switch u := updates.(type) {
	case *tg.UpdateShortSentMessage:
		return u.ID
	case *tg.Updates:
		for _, update := range u.Updates {
			// Verify if it's channel or group (Keeping, but not really necessary I guess)
			if msg, ok := update.(*tg.UpdateNewChannelMessage); ok {
				if m, ok := msg.Message.(*tg.Message); ok {
					return m.ID
				}
			}
			if msg, ok := update.(*tg.UpdateNewMessage); ok {
				if m, ok := msg.Message.(*tg.Message); ok {
					return m.ID
				}
			}
		}
	}
	return 0
}
