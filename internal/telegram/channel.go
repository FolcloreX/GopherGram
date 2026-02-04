package telegram

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

func (c *Client) UpdateChannelInfo(ctx context.Context, coverPath string, bio string) error {
	if c.TargetPeer == nil {
		return fmt.Errorf("TargetPeer nulo")
	}

	fmt.Println("⚙️  Atualizando Perfil do Canal/Grupo...")

	var photoInput tg.InputChatPhotoClass
	if coverPath != "" {
		fmt.Printf("   📸 Uploading avatar: %s...\n", filepath.Base(coverPath))
		file, err := c.uploader.FromPath(ctx, coverPath)
		if err != nil {
			return fmt.Errorf("erro upload avatar: %w", err)
		}
		photoInput = &tg.InputChatUploadedPhoto{File: file}
	}

	fmt.Print("   📝 Atualizando descrição... ")
	_, err := c.client.API().MessagesEditChatAbout(ctx, &tg.MessagesEditChatAboutRequest{
		Peer:  c.TargetPeer,
		About: bio,
	})
	if err != nil {
		fmt.Printf("Falha: %v\n", err)
	} else {
		fmt.Println("OK!")
	}

	if photoInput != nil {
		fmt.Print("   🖼 Atualizando foto... ")

		var errPhoto error

		switch p := c.TargetPeer.(type) {
		case *tg.InputPeerChannel:
			// Channels|Supergroups
			_, errPhoto = c.client.API().ChannelsEditPhoto(ctx, &tg.ChannelsEditPhotoRequest{
				Channel: &tg.InputChannel{
					ChannelID:  p.ChannelID,
					AccessHash: p.AccessHash,
				},
				Photo: photoInput,
			})
		case *tg.InputPeerChat:
			// For basic groups
			_, errPhoto = c.client.API().MessagesEditChatPhoto(ctx, &tg.MessagesEditChatPhotoRequest{
				ChatID: p.ChatID,
				Photo:  photoInput,
			})
		default:
			errPhoto = fmt.Errorf("tipo de chat não suporta foto")
		}

		if errPhoto != nil {
			fmt.Printf("Falha: %v\n", errPhoto)
		} else {
			fmt.Println("OK!")
		}
	}

	return nil
}

func (c *Client) CreateOriginChannel(ctx context.Context, title string) error {
	fmt.Printf("🆕 Criando novo canal de conteúdo: '%s'...\n", title)

	updates, err := c.client.API().ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
		Broadcast: true, // True = Channel, False = Grup
		Title:     title,
		About:     "Curso postado via GopherGram",
	})

	if err != nil {
		return fmt.Errorf("erro API CreateChannel: %w", err)
	}

	// Extract the channel object response
	var newChannel *tg.Channel

	switch u := updates.(type) {
	case *tg.Updates:
		for _, chat := range u.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				newChannel = ch
				break
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range u.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				newChannel = ch
				break
			}
		}
	}

	if newChannel == nil {
		return fmt.Errorf("canal criado, mas objeto de retorno é nulo")
	}

	// Update the config
	c.chatID = newChannel.ID
	c.TargetPeer = &tg.InputPeerChannel{
		ChannelID:  newChannel.ID,
		AccessHash: newChannel.AccessHash,
	}

	fmt.Printf("✅ Canal Criado com Sucesso! ID: %d\n", c.chatID)
	return nil
}

func (c *Client) resolvePeerByID(ctx context.Context, targetID int64) (tg.InputPeerClass, error) {
	fmt.Printf("   ↳ Consultando ID %d na API...\n", targetID)

	// Try channel/SuperGroup
	if chs, err := c.client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: targetID, AccessHash: 0},
	}); err == nil {
		var chatList []tg.ChatClass
		switch r := chs.(type) {
		case *tg.MessagesChats:
			chatList = r.Chats
		case *tg.MessagesChatsSlice:
			chatList = r.Chats
		}

		for _, chat := range chatList {
			if ch, ok := chat.(*tg.Channel); ok {
				fmt.Printf("      ✅ Encontrado Canal/Supergrupo: '%s'\n", ch.Title)
				return &tg.InputPeerChannel{ChannelID: ch.ID, AccessHash: ch.AccessHash}, nil
			}
		}
	}

	// Try as basic group
	if chats, err := c.client.API().MessagesGetChats(ctx, []int64{targetID}); err == nil {
		var chatList []tg.ChatClass
		switch r := chats.(type) {
		case *tg.MessagesChats:
			chatList = r.Chats
		case *tg.MessagesChatsSlice:
			chatList = r.Chats
		}

		for _, chat := range chatList {
			if ch, ok := chat.(*tg.Chat); ok {
				fmt.Printf("      ✅ Encontrado Grupo Básico: '%s'\n", ch.Title)
				return &tg.InputPeerChat{ChatID: ch.ID}, nil
			}
		}
	}

	return nil, fmt.Errorf("ID %d não encontrado ou sem acesso", targetID)
}

// Resolves the group where we will post the invite card
func (c *Client) ResolvePostTarget(ctx context.Context) error {
	if c.postGroupID == 0 {
		fmt.Println("📢 Nenhum POST_GROUP_ID definido. Convite vai para 'Saved Messages'.")
		c.PostTargetPeer = &tg.InputPeerSelf{}
		return nil
	}

	fmt.Println("🔄 Resolvendo Grupo de Divulgação...")
	peer, err := c.resolvePeerByID(ctx, c.postGroupID)
	if err != nil {
		return fmt.Errorf("erro ao resolver post group: %w", err)
	}
	c.PostTargetPeer = peer
	return nil
}

// GenerateInviteLink
func (c *Client) GenerateInviteLink(ctx context.Context) (string, error) {
	if c.TargetPeer == nil {
		return "", fmt.Errorf("TargetPeer nulo")
	}

	fmt.Println("🔗 Gerando link...")
	invite, err := c.client.API().MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
		Peer: c.TargetPeer, Title: "GopherGram",
	})
	if err != nil {
		return "", err
	}

	if exported, ok := invite.(*tg.ChatInviteExported); ok {
		return exported.Link, nil
	}
	return "", fmt.Errorf("link desconhecido")
}

func (c *Client) SendAnnouncement(
	ctx context.Context,
	coverPath string,
	caption string,
) error {

	if c.PostTargetPeer == nil {
		return fmt.Errorf("PostTargetPeer nulo")
	}

	fmt.Println("📢 Enviando Anúncio...")

	captionOpt := html.String(nil, caption)

	var inputMedia tg.InputMediaClass

	// The cover is optional
	if coverPath != "" {
		fmt.Printf("   📸 Com Imagem: %s\n", filepath.Base(coverPath))

		uploadedFile, err := c.uploader.FromPath(ctx, coverPath)
		if err != nil {
			return fmt.Errorf("erro upload capa: %w", err)
		}

		inputMedia = &tg.InputMediaUploadedPhoto{
			File: uploadedFile,
		}
	}

	baseReq := c.sender.To(c.PostTargetPeer)

	var sendErr error

	// Check wheter or not a topic was specifiec
	if c.postTopicID > 0 {
		fmt.Printf("   ↳ No Tópico ID: %d\n", c.postTopicID)

		topicReq := baseReq.Reply(c.postTopicID)

		if inputMedia != nil {
			_, sendErr = topicReq.Media(
				ctx,
				message.Media(inputMedia, captionOpt),
			)
		} else {
			_, sendErr = topicReq.StyledText(ctx, captionOpt)
		}

	} else {
		if inputMedia != nil {
			_, sendErr = baseReq.Media(
				ctx,
				message.Media(inputMedia, captionOpt),
			)
		} else {
			_, sendErr = baseReq.StyledText(ctx, captionOpt)
		}
	}

	if sendErr != nil {
		return fmt.Errorf("erro envio anúncio: %w", sendErr)
	}

	fmt.Println("✅ Anúncio postado com sucesso!")
	return nil
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
