package processor

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type VideoMeta struct {
	Width     int
	Height    int
	Duration  int
	ThumbPath string
}

func ExtractMetadata(videoPath string) (*VideoMeta, error) {
	// Gets the Widht, Height and Duration through ffprobe
	// Command: 
	// ffprobe 
	// -v error
	// -select_streams v:0 
	// -show_entries 
	// stream=width,height,duration -of csv=s=x:p=0 input.mp4
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration",
		"-of", "csv=s=x:p=0",
		videoPath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro no ffprobe: %w", err)
	}

	// Output example: "1920x1080x300.5"w
	parts := strings.Split(strings.TrimSpace(out.String()), "x")
	if len(parts) < 3 {
		return nil, fmt.Errorf("dados do ffprobe inválidos: %s", out.String())
	}

	width, _ := strconv.Atoi(parts[0])
	height, _ := strconv.Atoi(parts[1])
	durFloat, _ := strconv.ParseFloat(parts[2], 64)

	// Capture the frame in the first second of video to use as thumbnail
	thumbPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + "_thumb.jpg"

	// Command: ffmpeg 
	// -i input.mp4 
	// -ss 00:00:01
	// -vframes 1 
	// -vf scale=320:-1
	//  output.jpg -y
	// scale=320:-1 Ensure that the thumb is lightweight (Telegram wants < 200KB)
	cmdThumb := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", "00:00:01",
		"-vframes", "1",
		"-vf", "scale=320:-1",
		"-y",
		thumbPath,
	)

	if err := cmdThumb.Run(); err != nil {
		fmt.Printf("⚠️ Aviso: Não foi possível gerar thumbnail: %v\n", err)
		thumbPath = ""
	}

	return &VideoMeta{
		Width:     width,
		Height:    height,
		Duration:  int(durFloat),
		ThumbPath: thumbPath,
	}, nil
}
