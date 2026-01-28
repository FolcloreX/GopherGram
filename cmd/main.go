package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/FolcloreX/GopherGram/internal/config"
	domain "github.com/FolcloreX/GopherGram/internal/domains"
	"github.com/FolcloreX/GopherGram/internal/processor"
	"github.com/FolcloreX/GopherGram/internal/scanner"
	"github.com/FolcloreX/GopherGram/internal/telegram"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("❌ Uso correto: go run cmd/bot/main.go <caminho_absoluto_da_pasta>")
		fmt.Println("Exemplo: go run cmd/bot/main.go /home/user/meu_curso")
		os.Exit(1)
	}

	rawPath := os.Args[1]
	rootDir, err := filepath.Abs(rawPath)
	if err != nil {
		log.Fatalf("Erro ao resolver caminho absoluto: %v", err)
	}

	info, err := os.Stat(rootDir)
	if os.IsNotExist(err) {
		log.Fatalf("O diretório informado não existe: %s", rootDir)
	}
	if !info.IsDir() {
		log.Fatalf("O caminho informado não é um diretório: %s", rootDir)
	}

	fmt.Printf("📂 Diretório Base Definido: %s\n", rootDir)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro config: %v", err)
	}

	bot := telegram.NewClient(cfg)

	err = bot.Start(context.Background(), func(ctx context.Context) error {
		if err := bot.CheckChatAccess(ctx); err != nil {
			return err
		}

		fmt.Printf("🔍 Escaneando estrutura em: %s\n", rootDir)
		scan := scanner.New(rootDir)
		course, err := scan.Scan()
		if err != nil {
			return err
		}

		fmt.Printf("📦 Conteúdo: %d Módulos, %d Assets soltos\n", len(course.Modules), len(course.Assets))

		if len(course.Assets) > 0 {
			fmt.Println("\n📂 [1/3] Processando Arquivos de Apoio...")

			// The zip.file will be created here, not in the target folder.
			// Be aware of that
			zipName := "files.zip"

			zipper := processor.Zipper{RootDir: rootDir}

			if err := zipper.ZipFiles(course.Assets, zipName); err != nil {
				return fmt.Errorf("erro ao zipar: %w", err)
			}

			parts, err := processor.SplitFileBinary(zipName, domain.MaxFileSize)
			if err != nil {
				return fmt.Errorf("erro ao dividir zip: %w", err)
			}

			for _, part := range parts {
				caption := fmt.Sprintf("🗂 <b>Material de Apoio</b>\nArquivo: %s", part)

				if err := bot.UploadAndSendDocument(ctx, part, caption); err != nil {
					return err
				}
				os.Remove(part)
			}
			os.Remove(zipName)
		}

		fmt.Println("\n🎬 [2/3] Processando Vídeos...")
		var indexBuilder strings.Builder

		indexBuilder.WriteString("⚠️ <b>Menu do Curso</b> ⚠️\n\n")

		ffmpeg := &processor.FFmpegSplitter{}

		for _, mod := range course.Modules {
			fmt.Printf("🔹 Módulo: %s\n", mod.Name)
			indexBuilder.WriteString(fmt.Sprintf("\n📁 <b>%s</b>\n", mod.Name))

			for _, video := range mod.Videos {
				parts, err := ffmpeg.SplitVideo(video.FilePath, domain.MaxFileSize)
				if err != nil {
					log.Printf("Erro split vídeo %s: %v", video.FileName, err)
					continue
				}

				for i, partPath := range parts {
					caption := video.FormatCaption()
					if len(parts) > 1 {
						caption += fmt.Sprintf(" [Parte %d/%d]", i+1, len(parts))
					}

					if err := bot.UploadAndSendVideo(ctx, partPath, caption); err != nil {
						return err
					}

					if partPath != video.FilePath {
						os.Remove(partPath)
					}
				}
				indexBuilder.WriteString(fmt.Sprintf("#%s ", video.ID))
			}
			indexBuilder.WriteString("\n")
		}

		fmt.Println("\n📑 [3/3] Finalizando e Pinando...")

		msgID, err := bot.SendMessage(ctx, indexBuilder.String())
		if err != nil {
			log.Printf("Erro envio Index: %v", err)
		} else {
			if err := bot.PinMessage(ctx, msgID); err != nil {
				log.Printf("Aviso: Não consegui pinar a mensagem: %v", err)
			} else {
				fmt.Println("📌 Index fixado!")
			}
		}

		fmt.Println("\n🚀 Upload Completo!")
		return nil
	})

	if err != nil {
		log.Fatalf("Falha crítica na execução: %v", err)
	}
}
