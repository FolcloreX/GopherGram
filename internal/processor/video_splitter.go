package processor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type FFmpegSplitter struct{}

// SplitVideo divide the files if it's exceed the limit 2GB.
// Return the list of generated files or original if not splitted.
func (fs *FFmpegSplitter) SplitVideo(inputFile string, limitBytes int64) ([]string, error) {
	info, err := os.Stat(inputFile)
	if err != nil {
		return nil, err
	}

	safeLimit := int64(float64(limitBytes) * 0.95)

	if info.Size() <= safeLimit {
		return []string{inputFile}, nil
	}

	fmt.Printf("🔄 Dividindo vídeo grande: %s (%d bytes)\n", filepath.Base(inputFile), info.Size())

	durationSec, err := getVideoDuration(inputFile)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter duração: %w", err)
	}

	// Ex: 5GB file, 2GB limit -> 3 parts
	numParts := int(info.Size()/safeLimit) + 1
	segmentTime := durationSec / float64(numParts)

	// output_%03d.mp4 generates output_000.mp4, output_001.mp4...
	ext := filepath.Ext(inputFile)
	baseName := strings.TrimSuffix(filepath.Base(inputFile), ext)
	outputPattern := filepath.Join(filepath.Dir(inputFile), fmt.Sprintf("%s-part-%%1d%s", baseName, ext))

	// Command FFmpeg:
	// -c copy: Copy streams (fast, no re-encode)
	// -f segment: Use segments muxer
	// -reset_timestamps 1: Each part will be "touchable" individually
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-c", "copy",
		"-map", "0",
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%.2f", segmentTime),
		"-reset_timestamps", "1",
		outputPattern,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg falhou: %s | Log: %s", err, stderr.String())
	}

	// TODO, improve this mess
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(inputFile), fmt.Sprintf("%s-part-*%s", baseName, ext)))
	if err != nil {
		return nil, err
	}

	return matches, nil
}

// getVideoDuration uses ffprobe to get the exact duration
func getVideoDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	val := strings.TrimSpace(string(out))
	return strconv.ParseFloat(val, 64)
}
