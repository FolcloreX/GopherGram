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
	"github.com/FolcloreX/GopherGram/internal/telegram"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("\n❌ ERRO: Você precisa informar a pasta do curso!")
		fmt.Println("---------------------------------------------------------")
		fmt.Println("✅ Uso Correto:")
		fmt.Println("   go run cmd/bot/main.go \"/caminho/completo/do/curso\"")
		fmt.Println("---------------------------------------------------------")
		os.Exit(1)
	}

	rawPath := os.Args[1]
	rootDir, err := filepath.Abs(rawPath)
	if err != nil {
		log.Fatalf("Erro ao ler caminho: %v", err)
	}

	info, err := os.Stat(rootDir)
	if os.IsNotExist(err) {
		log.Fatalf("❌ A pasta informada não existe:\n%s", rootDir)
	}
	if !info.IsDir() {
		log.Fatalf("❌ O caminho informado é um arquivo, não uma pasta:\n%s", rootDir)
	}

	fmt.Printf("\n🚀 INICIANDO GOPHERGRAM UPLOADER\n")
	fmt.Printf("📂 Alvo: %s\n\n", rootDir)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro no .env: %v", err)
	}

	bot := telegram.NewClient(cfg)

	err = bot.Start(context.Background(), func(ctx context.Context) error {

		if err := bot.CheckChatAccess(ctx); err != nil {
			return err
		}

		fmt.Println("🔍 Escaneando arquivos...")
		scan := scanner.New(rootDir)
		course, err := scan.Scan()
		if err != nil {
			return fmt.Errorf("erro no scanner: %w", err)
		}

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
				// Generates the tag (Ex: #Doc001, #Doc002...)
				currentDocTag := fmt.Sprintf("#%s%03d", domain.HashTagDoc, i+1)
				indexBuilder.WriteString(currentDocTag + " ")

				caption := fmt.Sprintf("%s 🗂 <b>Material de Apoio</b>\nArquivo %d/%d",
					currentDocTag, i+1, len(parts))

				if err := bot.UploadAndSendDocument(ctx, part, caption); err != nil {
					return err
				}
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

						if partPath != video.FilePath {
							os.Remove(partPath)
						}
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
			fmt.Println("📑 [FASE 3] Enviando Menu Final")
			fmt.Println("------------------------------------------------")

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

		fmt.Println("\n✅✅✅ PROCESSO CONCLUÍDO COM SUCESSO! ✅✅✅")
		return nil
	})

	if err != nil {
		fmt.Printf("\n❌ FALHA CRÍTICA: %v\n", err)
		os.Exit(1)
	}
}
