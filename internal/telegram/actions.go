package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FolcloreX/GopherGram/internal/processor"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/schollz/progressbar/v3"
)


type progressWrapper struct {
	bar *progressbar.ProgressBar
}

func (p *progressWrapper) Chunk(ctx context.Context, state uploader.ProgressState) error {
	_ = p.bar.Set64(state.Uploaded)
	return nil
}

func (c *Client) UploadAndSendVideo(ctx context.Context, filePath string, caption string, meta *processor.VideoMeta) error {
	fileName := filepath.Base(filePath)

	if c.TargetPeer == nil {
		return fmt.Errorf("TargetPeer nulo (rode CheckChatAccess)")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("erro ler arquivo: %w", err)
	}

	fmt.Printf("\n🎬 Enviando: %s [%dx%d]\n", fileName, meta.Width, meta.Height)

	bar := progressbar.DefaultBytes(info.Size(), "⬆️  Video")
	u := c.uploader.WithProgress(&progressWrapper{bar: bar})

	videoUpload, err := u.FromPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("falha upload video: %w", err)
	}

	_ = bar.Finish()
	fmt.Println()

	var thumbUpload tg.InputFileClass
	if meta.ThumbPath != "" {
		fmt.Print("🖼 Enviando thumbnail... ")
		if t, err := c.uploader.FromPath(ctx, meta.ThumbPath); err == nil {
			thumbUpload = t
			fmt.Println("OK")
		} else {
			fmt.Println("Falhou (ignorando)")
		}
	}

	attrs := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeVideo{
			SupportsStreaming: true,
			Duration:          float64(meta.Duration),
			W:                 meta.Width,
			H:                 meta.Height,
		},
		&tg.DocumentAttributeFilename{FileName: fileName},
	}

	inputMedia := &tg.InputMediaUploadedDocument{
		File:       videoUpload,
		MimeType:   "video/mp4",
		Attributes: attrs,
		ForceFile:  false,
	}

	if thumbUpload != nil {
		inputMedia.Thumb = thumbUpload
	}

	fmt.Print("📨 Enviando mensagem da mídia... ")

	_, err = c.sender.
		To(c.TargetPeer).
		Media(
			ctx,
			message.Media(inputMedia, html.String(nil, caption)),
		)

	if err != nil {
		return fmt.Errorf("erro envio: %w", err)
	}

	return nil
}

func (c *Client) UploadAndSendDocument(ctx context.Context, filePath string, caption string) error {
	fileName := filepath.Base(filePath)

	if c.TargetPeer == nil {
		return fmt.Errorf("TargetPeer nulo")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("erro ler arquivo: %w", err)
	}

	fmt.Printf("\n🗂 Enviando: %s\n", fileName)

	bar := progressbar.DefaultBytes(info.Size(), "⬆️  Doc  ")
	u := c.uploader.WithProgress(&progressWrapper{bar: bar})

	fileUpload, err := u.FromPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("falha upload doc: %w", err)
	}

	_ = bar.Finish()
	fmt.Println()

	attrs := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeFilename{FileName: fileName},
	}

	inputMedia := &tg.InputMediaUploadedDocument{
		File:       fileUpload,
		MimeType:   "application/zip", // Or application/octet-stream
		Attributes: attrs,
		ForceFile:  true,
	}

	fmt.Print("📨 Enviando Mensagem dos Materiais Zipados... ")

	_, err = c.sender.
		To(c.TargetPeer).
		Media(
			ctx,
			message.Media(inputMedia, html.String(nil, caption)),
		)

	if err != nil {
		return fmt.Errorf("erro envio: %w", err)
	}

	return nil
}

func (c *Client) SendMessage(ctx context.Context, text string) (int, error) {
	if c.TargetPeer == nil {
		return 0, fmt.Errorf("TargetPeer nulo")
	}
	updates, err := c.sender.To(c.TargetPeer).StyledText(ctx, html.String(nil, text))
	if err != nil {
		return 0, err
	}
	return extractMsgID(updates), nil
}

func (c *Client) PinMessage(ctx context.Context, messageID int) error {
	if c.TargetPeer == nil {
		return fmt.Errorf("TargetPeer nulo")
	}
	_, err := c.client.API().MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer: c.TargetPeer, ID: messageID, Silent: true,
	})
	return err
}

func (c *Client) CheckChatAccess(ctx context.Context) error {
	fmt.Printf("🔄 Resolvendo ID %d...\n", c.chatID)

	// Check for channel
	if chs, err := c.client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: c.chatID, AccessHash: 0},
	}); err == nil {
		switch r := chs.(type) {
		case *tg.MessagesChats:
			if len(r.Chats) > 0 {
				if ch, ok := r.Chats[0].(*tg.Channel); ok {
					c.TargetPeer = &tg.InputPeerChannel{ChannelID: ch.ID, AccessHash: ch.AccessHash}
					fmt.Printf("✅ Canal: %s\n", ch.Title)
					return nil
				}
			}
		case *tg.MessagesChatsSlice:
			if len(r.Chats) > 0 {
				if ch, ok := r.Chats[0].(*tg.Channel); ok {
					c.TargetPeer = &tg.InputPeerChannel{ChannelID: ch.ID, AccessHash: ch.AccessHash}
					fmt.Printf("✅ Canal: %s\n", ch.Title)
					return nil
				}
			}
		}
	}

	// Check for group
	if chats, err := c.client.API().MessagesGetChats(ctx, []int64{c.chatID}); err == nil {
		switch r := chats.(type) {
		case *tg.MessagesChats:
			if len(r.Chats) > 0 {
				if ch, ok := r.Chats[0].(*tg.Chat); ok {
					c.TargetPeer = &tg.InputPeerChat{ChatID: ch.ID}
					fmt.Printf("✅ Grupo: %s\n", ch.Title)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("ID %d não encontrado", c.chatID)
}

func extractMsgID(updates tg.UpdatesClass) int {
	switch u := updates.(type) {
	case *tg.UpdateShortSentMessage:
		return u.ID
	case *tg.Updates:
		for _, update := range u.Updates {
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
