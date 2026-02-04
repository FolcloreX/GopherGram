package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/FolcloreX/GopherGram/internal/processor"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/schollz/progressbar/v3"
)

type progressWrapper struct {
	bar *progressbar.ProgressBar
}

func (p *progressWrapper) Chunk(ctx context.Context, state uploader.ProgressState) error {
	_ = p.bar.Set64(state.Uploaded)
	return nil
}

func (c *Client) uploadWithRetry(
	ctx context.Context,
	filePath string,
	label string,
	timeout time.Duration,
) (tg.InputFileClass, error) {
	fileName := filepath.Base(filePath)
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("erro crítico ao abrir arquivo: %w", err)
		}

		info, _ := file.Stat()
		fmt.Printf("\n%s %s (Tentativa %d/%d)\n", label, fileName, attempt, maxRetries)

		bar := progressbar.DefaultBytes(info.Size(), label)
		u := c.uploader.WithProgress(&progressWrapper{bar: bar})

		uploadCtx, cancel := context.WithTimeout(ctx, timeout)

		uploadObj := uploader.NewUpload(fileName, file, info.Size())

		inputFile, err := u.Upload(uploadCtx, uploadObj)

		file.Close()
		cancel()
		_ = bar.Finish()
		fmt.Println()

		if err == nil {
			return inputFile, nil
		}

		if d, ok := tgerr.AsFloodWait(err); ok {
			fmt.Printf("⏳ FloodWait detectado. Aguardando %v...\n", d)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(d + 1*time.Second):
				continue
			}
		}

		fmt.Printf("⚠️  Falha no upload: %v\n", err)
		if attempt < maxRetries {
			time.Sleep(3 * time.Second)
			continue
		}

		return nil, fmt.Errorf("desistindo após %d tentativas: %w", maxRetries, err)
	}

	return nil, fmt.Errorf("falha desconhecida no loop de upload")
}

func (c *Client) UploadAndSendVideo(
	ctx context.Context,
	filePath string,
	caption string,
	meta *processor.VideoMeta,
) error {
	if c.TargetPeer == nil {
		return fmt.Errorf("TargetPeer nulo (rode CheckChatAccess)")
	}

	fileName := filepath.Base(filePath)
	fmt.Printf("\n🎬 Enviando vídeo: %s [%dx%d]\n", fileName, meta.Width, meta.Height)

	// File upload with retry
	videoUpload, err := c.uploadWithRetry(ctx, filePath, "⬆️  Video", 120*time.Minute)
	if err != nil {
		return err
	}

	// Upload of the thumbanil (Best-effort)
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

	// 3. Montagem dos Atributos
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

	fmt.Print("📨 Enviando mensagem do vídeo... ")
	_, err = c.sender.
		To(c.TargetPeer).
		Media(ctx, message.Media(inputMedia, html.String(nil, caption)))

	// Treating floodAwait
	if err != nil {
		if d, ok := tgerr.AsFloodWait(err); ok {
			fmt.Printf("\n⏳ FloodWait no envio (%v). Aguardando...\n", d)
			time.Sleep(d + 1*time.Second)
			return c.UploadAndSendVideo(ctx, filePath, caption, meta) // Recursive Retry
		}
		return fmt.Errorf("erro envio vídeo: %w", err)
	}

	fmt.Println("✅ Vídeo enviado com sucesso")
	return nil
}

func (c *Client) UploadAndSendDocument(
	ctx context.Context,
	filePath string,
	caption string,
) error {
	if c.TargetPeer == nil {
		return fmt.Errorf("TargetPeer nulo")
	}

	fileName := filepath.Base(filePath)
	fmt.Printf("\n🗂 Processando: %s\n", fileName)

	fileUpload, err := c.uploadWithRetry(ctx, filePath, "⬆️  Doc  ", 60*time.Minute)
	if err != nil {
		return err
	}

	attrs := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeFilename{FileName: fileName},
	}

	inputMedia := &tg.InputMediaUploadedDocument{
		File:       fileUpload,
		MimeType:   "application/zip",
		Attributes: attrs,
		ForceFile:  true,
	}

	fmt.Print("📨 Enviando mensagem do documento... ")
	_, err = c.sender.
		To(c.TargetPeer).
		Media(ctx, message.Media(inputMedia, html.String(nil, caption)))

	if err != nil {
		if d, ok := tgerr.AsFloodWait(err); ok {
			fmt.Printf("\n⏳ FloodWait no envio (%v). Aguardando...\n", d)
			time.Sleep(d + 1*time.Second)
			return c.UploadAndSendDocument(ctx, filePath, caption)
		}
		return fmt.Errorf("erro envio doc: %w", err)
	}

	fmt.Println("✅ Documento enviado com sucesso")
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
