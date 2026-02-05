package processor

import (
	"fmt"
	"os"

	"github.com/FolcloreX/GopherGram/internal/domain"
)

func CalculateAssetsSize(assets []string) int64 {
	var total int64
	for _, path := range assets {
		if info, err := os.Stat(path); err == nil {
			total += info.Size()
		}
	}
	return total
}

func CalculateVideosSize(modules []*domain.Module) int64 {
	var total int64
	for _, mod := range modules {
		for _, vid := range mod.Videos {
			total += vid.Size
		}
	}
	return total
}

func FormatCourseCard(courseName string, totalBytes int64, totalSeconds int, logo string, inviteLink string) string {
	gb := float64(totalBytes) / (1024 * 1024 * 1024)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60

	buttonHTML := ""
	if inviteLink != "" {
		buttonHTML = fmt.Sprintf("\n\n👉 <a href=\"%s\"><b>CLIQUE PARA ACESSAR</b></a> 👈", inviteLink)
	}

	return fmt.Sprintf(
		"🎓 <b>%s</b>\n\n"+
			"💾 | Tamanho Total: %.2f GB\n"+
			"⏳ | Duração Total: %dh %02dm\n"+
			"🚀 | Lançamento: 2024\n\n"+
			"%s\n\n"+
			"%s",
		courseName,
		gb,
		hours,
		minutes,
		logo,
		buttonHTML,
	)
}

func FormatChannelBio(totalBytes int64, totalSeconds int, inviteLink, logo string) string {
	gb := float64(totalBytes) / (1024 * 1024 * 1024)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60

	return fmt.Sprintf(
		"Tamanho: %.2f GB\n"+
			"Duração: %dh %02dm\n"+
			"Convite: %s\n\n"+
			"%s",
		gb, hours, minutes, inviteLink, logo,
	)
}
