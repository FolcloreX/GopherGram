package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/FolcloreX/GopherGram/internal/config"
	"github.com/FolcloreX/GopherGram/internal/domain"
	"github.com/FolcloreX/GopherGram/internal/processor"
	"github.com/FolcloreX/GopherGram/internal/scanner"
	"github.com/FolcloreX/GopherGram/internal/state"
	"github.com/FolcloreX/GopherGram/internal/telegram"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("\n❌ ERRO: Informe a pasta do curso!")
		fmt.Println("✅ Uso: go run cmd/bot/main.go \"/path/curso\" \"/path/capa.jpg\"(opcional)")
		os.Exit(1)
	}

	// Argument 1: Content Directory Absolute Path
	rawPath := os.Args[1]
	rootDir, err := filepath.Abs(rawPath)
	if err != nil {
		log.Fatalf("Erro path: %v", err)
	}

	info, err := os.Stat(rootDir)
	if os.IsNotExist(err) || !info.IsDir() {
		log.Fatalf("Pasta inválida: %s", rootDir)
	}

	// Argument 2: Cover Absolute Path (Opcional)
	coverPath := ""
	if len(os.Args) >= 3 {
		coverPath = os.Args[2]
		if _, err := os.Stat(coverPath); os.IsNotExist(err) {
			log.Printf("⚠️ Aviso: Capa informada não existe: %s (será enviado como texto)", coverPath)
			coverPath = ""
		}
	}

	// ContentName = Name of the Root Folder
	contentName := filepath.Base(rootDir)

	fmt.Printf("\n🚀 GOPHERGRAM UPLOADER\n📂 Curso: %s\n🖼 Capa: %s\n\n", contentName, coverPath)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro no .env: %v", err)
	}

	bot := telegram.NewClient(cfg)

	prog, err := state.LoadProgressContent(contentName)

	if err != nil {
		log.Printf("⚠️  Erro state: %v", err)
		// Fallback se der erro crítico de IO
		prog, _ = state.Load("session/progress_fallback.json")
	} else {
		fmt.Printf("💾 Estado carregado: %s\n", prog.FilePath)
	}

	err = bot.Start(context.Background(), func(ctx context.Context) error {
		// Check if a group was passed otherwise create a new one
		if cfg.ChatID != 0 {
			if err := bot.CheckChatAccess(ctx); err != nil {
				return fmt.Errorf("erro ao acessar chat origem: %w", err)
			}
		} else {
			if err := bot.CreateOriginChannel(ctx, contentName); err != nil {
				return fmt.Errorf("erro ao criar canal: %w", err)
			}
		}

		// Resolve the PostGroup, if no one is passed we send to the saved messages
		if err := bot.ResolvePostTarget(ctx); err != nil {
			return fmt.Errorf("erro ao resolver grupo de divulgação: %w", err)
		}

		fmt.Println("🔍 Escaneando arquivos...")
		scan := scanner.New(rootDir)
		course, err := scan.Scan()
		if err != nil {
			return fmt.Errorf("erro no scanner: %w", err)
		}

		// Metadada to perform information in the invite and the description
		var totalSizeBytes int64
		var totalDurationSeconds int

		totalSizeBytes = processor.CalculateAssetsSize(course.Assets)
		totalSizeBytes += processor.CalculateVideosSize(course.Modules)

		totalVideos := 0
		for _, m := range course.Modules {
			totalVideos += len(m.Videos)
		}

		fmt.Printf("📊 Resumo: %d Módulos | %d Vídeos | %d Assets (Arquivos Extra)\n",
			len(course.Modules), totalVideos, len(course.Assets))

		if totalVideos == 0 && len(course.Assets) == 0 {
			fmt.Println("⚠️  Nenhum arquivo encontrado! Verifique se a pasta está correta.")
			return nil
		}

		// Create the menu index
		var indexBuilder strings.Builder
		indexBuilder.WriteString("⚠️ <b>Menu do Curso</b> ⚠️\n\n")
		indexBuilder.WriteString("Clique nas hashtags para navegar.\n\n")
		indexBuilder.WriteString("📂 <b>Arquivos</b>\n")

		if len(course.Assets) > 0 {
			fmt.Println("\n------------------------------------------------")
			fmt.Println("📂 [FASE 1] Processando Material de Apoio")
			fmt.Println("------------------------------------------------")

			zipName := "Arquivos.zip"
			zipper := processor.Zipper{RootDir: rootDir}

			if err := zipper.ZipFiles(course.Assets, zipName); err != nil {
				return err
			}

			parts, err := processor.SplitFileBinary(zipName, domain.MaxFileSize)
			if err != nil {
				return err
			}

			for i, part := range parts {

				// Check if the part is not already uploaded
				if prog.IsDone(part) {
					fmt.Printf("⏩ Pulando (já enviado): %s\n", filepath.Base(part))
					os.Remove(part) // Remove because we already sent
					continue
				}

				// Generates the tag (Ex: #Doc001, #Doc002...)
				currentDocTag := fmt.Sprintf("#%s%03d", domain.HashTagDoc, i+1)
				indexBuilder.WriteString(currentDocTag + " ")

				caption := fmt.Sprintf("%s 🗂 <b>Material de Apoio</b>\nArquivo %d/%d",
					currentDocTag, i+1, len(parts))

				if err := bot.UploadAndSendDocument(ctx, part, caption); err != nil {
					return err
				}

				// If uploaded sucessfully we mark as done in the state
				prog.MarkAsDone(part)
				os.Remove(part)
			}
			indexBuilder.WriteString("\n\n")

			os.Remove(zipName)

		} else {
			indexBuilder.WriteString("<i>Nenhum material de apoio.</i>\n\n")
		}

		if totalVideos > 0 {
			fmt.Println("\n------------------------------------------------")
			fmt.Println("🎬 [FASE 2] Processando Vídeos")
			fmt.Println("------------------------------------------------")

			ffmpeg := &processor.FFmpegSplitter{}

			for _, mod := range course.Modules {
				fmt.Printf("\n🔹 Processando Módulo: %s\n", mod.Name)
				indexBuilder.WriteString(fmt.Sprintf("\n📁 <b>%s</b>\n", mod.Name))

				for _, video := range mod.Videos {
					// Split the video if greater than 2GB
					parts, err := ffmpeg.SplitVideo(video.FilePath, domain.MaxFileSize)
					if err != nil {
						log.Printf("❌ Erro ao dividir vídeo %s: %v", video.FileName, err)
						continue
					}

					for i, partPath := range parts {
						// Extracting metada
						fmt.Printf("   📸 Gerando metadados para %s...\n", filepath.Base(partPath))
						meta, err := processor.ExtractMetadata(partPath)

						// TODO improve the logic! Code repetition
						// Sum the total duration to show
						// We keep it before check if it's uploaded or not to keep the metadada correct
						totalDurationSeconds += meta.Duration

						if prog.IsDone(partPath) {
							fmt.Printf("⏩ Pulando (já enviado): %s\n", filepath.Base(partPath))

							if partPath != video.FilePath {
								os.Remove(partPath)
							}
							continue
						}

						if err != nil {
							log.Printf("   ⚠️ Falha ao gerar thumbnail (enviando sem): %v", err)
							meta = &processor.VideoMeta{}
						}

						caption := video.FormatCaption()
						if len(parts) > 1 {
							caption += fmt.Sprintf(" [Parte %d/%d]", i+1, len(parts))
						}

						if err := bot.UploadAndSendVideo(ctx, partPath, caption, meta); err != nil {
							return err
						}

						// Mark the file as correctly uploaded in the state
						prog.MarkAsDone(partPath)

						// Cleaning the generated chunks of file. ALWAYS keeping the original
						if partPath != video.FilePath {
							os.Remove(partPath)
						}

						// Clear the generate thumb
						if meta.ThumbPath != "" {
							os.Remove(meta.ThumbPath)
						}
					}
					// Add ID to the index
					indexBuilder.WriteString(fmt.Sprintf("#%s ", video.ID))
				}
				indexBuilder.WriteString("\n")
			}

			fmt.Println("\n------------------------------------------------")
			fmt.Println("📑 [FASE 3] Finalização")
			fmt.Println("------------------------------------------------")

			fmt.Print("📨 Enviando Menu de Links... ")
			msgID, err := bot.SendMessage(ctx, indexBuilder.String())
			if err != nil {
				log.Printf("Erro ao enviar Index: %v", err)
			} else {
				if err := bot.PinMessage(ctx, msgID); err != nil {
					log.Printf("Aviso: Falha ao pinar mensagem: %v", err)
				} else {
					fmt.Println("📌 Menu fixado com sucesso!")
				}
			}
		}

		fmt.Println("\n------------------------------------------------")
		fmt.Println("📢 [FASE 4] Divulgação")
		fmt.Println("------------------------------------------------")

		inviteLink, err := bot.GenerateInviteLink(ctx)
		if err != nil {
			log.Printf("⚠️ Erro link convite: %v", err)
			inviteLink = ""
		} else {
			fmt.Printf("🔗 Link do Curso: %s\n", inviteLink)
		}

		// Update the channel description
		channelBio := processor.FormatChannelBio(
			totalSizeBytes,
			totalDurationSeconds,
			inviteLink,
			cfg.Logo,
		)

		fmt.Println("\n🎨 Personalizando Canal...")
		if err := bot.UpdateChannelInfo(ctx, coverPath, channelBio); err != nil {
			log.Printf("⚠️ Erro ao atualizar perfil: %v", err)
		}

		// Create the invite message
		cardCaption := processor.FormatCourseCard(
			contentName,
			totalSizeBytes,
			totalDurationSeconds,
			cfg.Logo,
			inviteLink,
		)

		if err := bot.SendAnnouncement(ctx, coverPath, cardCaption); err != nil {
			log.Printf("❌ Erro ao postar anúncio: %v", err)
		}

		fmt.Println("\n✅✅✅ PROCESSO CONCLUÍDO! ✅✅✅")
		return nil
	})

	if err != nil {
		fmt.Printf("\n❌ FALHA: %v\n", err)
		os.Exit(1)
	}
}
