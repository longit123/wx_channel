package services

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yapingcat/gomedia/go-mp4"

	"wx_channel/internal/utils"
)

// 支持的音频格式
const (
	AudioFormatM4A = "m4a"
	AudioFormatMP3 = "mp3"
)

// AudioCapabilities 表示音频提取能力
type AudioCapabilities struct {
	FFmpegAvailable bool     `json:"ffmpegAvailable"`
	Formats         []string `json:"formats"`
}

// AudioExtractor 负责从 mp4 视频中提取音频轨
//
// 策略：
//   - m4a: 优先用 ffmpeg -acodec copy（无损、极快）；无 ffmpeg 时用纯 Go (gomedia) demux+mux
//   - mp3: 必须用 ffmpeg 转码（libmp3lame）
type AudioExtractor struct {
	ffmpegPath string // 缓存的 ffmpeg 可执行文件路径；空表示未安装
}

// NewAudioExtractor 创建音频提取器，并探测系统 ffmpeg
func NewAudioExtractor() *AudioExtractor {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		utils.Info("未检测到 ffmpeg，音频提取将仅支持 m4a 格式（纯 Go 抽轨）")
		return &AudioExtractor{ffmpegPath: ""}
	}
	utils.Info("✓ 检测到 ffmpeg: %s，音频提取支持 m4a 和 mp3 格式", path)
	return &AudioExtractor{ffmpegPath: path}
}

// Capabilities 返回当前环境的音频提取能力
func (a *AudioExtractor) Capabilities() AudioCapabilities {
	caps := AudioCapabilities{
		FFmpegAvailable: a.ffmpegPath != "",
		Formats:         []string{AudioFormatM4A},
	}
	if a.ffmpegPath != "" {
		caps.Formats = append(caps.Formats, AudioFormatMP3)
	}
	return caps
}

// Extract 从 mp4 视频提取音频并保存到 outputDir
//
// 参数：
//   - mp4Path: 源视频文件绝对路径
//   - outputDir: 输出目录（音频与视频同目录）
//   - format: "m4a" 或 "mp3"
//
// 返回生成的音频文件绝对路径
func (a *AudioExtractor) Extract(mp4Path, outputDir, format string) (string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = AudioFormatM4A
	}

	if format != AudioFormatM4A && format != AudioFormatMP3 {
		return "", fmt.Errorf("不支持的音频格式: %s (仅支持 m4a/mp3)", format)
	}

	// 校验源文件
	if _, err := os.Stat(mp4Path); err != nil {
		return "", fmt.Errorf("源视频文件不存在: %w", err)
	}

	// mp3 必须有 ffmpeg
	if format == AudioFormatMP3 && a.ffmpegPath == "" {
		return "", fmt.Errorf("mp3 格式需要 ffmpeg，但系统未安装；请改用 m4a 格式或安装 ffmpeg")
	}

	// 生成输出文件名：与视频同名换扩展名
	videoBase := filepath.Base(mp4Path)
	audioBase := strings.TrimSuffix(videoBase, filepath.Ext(videoBase)) + "." + format
	outputPath := utils.GenerateUniquePath(outputDir, audioBase)

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	start := time.Now()
	var err error
	switch format {
	case AudioFormatM4A:
		err = a.extractM4A(mp4Path, outputPath)
	case AudioFormatMP3:
		err = a.extractMP3(mp4Path, outputPath)
	}
	if err != nil {
		// 失败时清理可能产生的部分文件
		_ = os.Remove(outputPath)
		return "", err
	}

	utils.Info("✓ 音频提取完成: %s -> %s (耗时 %s)", filepath.Base(mp4Path), filepath.Base(outputPath), time.Since(start).Round(time.Millisecond))
	return outputPath, nil
}

// extractM4A 提取 m4a 音频
// 优先用 ffmpeg（-acodec copy，无损且极快）；无 ffmpeg 时用纯 Go gomedia
func (a *AudioExtractor) extractM4A(mp4Path, outputPath string) error {
	if a.ffmpegPath != "" {
		return a.runFFmpeg(mp4Path, outputPath, "-vn", "-acodec", "copy")
	}
	return extractM4APureGo(mp4Path, outputPath)
}

// extractMP3 用 ffmpeg 转码为 mp3
func (a *AudioExtractor) extractMP3(mp4Path, outputPath string) error {
	return a.runFFmpeg(mp4Path, outputPath, "-vn", "-acodec", "libmp3lame", "-ab", "192k")
}

// runFFmpeg 执行 ffmpeg 子进程
func (a *AudioExtractor) runFFmpeg(input, output string, args ...string) error {
	fullArgs := append([]string{"-y", "-i", input}, args...)
	fullArgs = append(fullArgs, output)

	cmd := exec.Command(a.ffmpegPath, fullArgs...)
	// 抑制 ffmpeg 的 stderr 噪音（除非出错）
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg 执行失败: %w", err)
	}
	return nil
}

// extractM4APureGo 用纯 Go (gomedia) 实现 m4a 抽轨
// demux 源 mp4 → 过滤 AAC 音频包 → mux 为只含音频的 mp4 (即 m4a)
func extractM4APureGo(mp4Path, outputPath string) error {
	srcFile, err := os.Open(mp4Path)
	if err != nil {
		return fmt.Errorf("打开源视频失败: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer dstFile.Close()

	demuxer := mp4.CreateMp4Demuxer(srcFile)
	if _, err := demuxer.ReadHead(); err != nil && err != io.EOF {
		return fmt.Errorf("读取 mp4 头部失败: %w", err)
	}

	muxer, err := mp4.CreateMp4Muxer(dstFile)
	if err != nil {
		return fmt.Errorf("创建 m4a muxer 失败: %w", err)
	}

	var atid uint32
	hasAudio := false
	packetCount := 0

	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			break
		}
		// 仅处理 AAC 音频包；MP3 音频也支持
		if pkg.Cid == mp4.MP4_CODEC_AAC || pkg.Cid == mp4.MP4_CODEC_MP3 {
			if !hasAudio {
				atid = muxer.AddAudioTrack(pkg.Cid)
				hasAudio = true
			}
			if err := muxer.Write(atid, pkg.Data, uint64(pkg.Pts), uint64(pkg.Dts)); err != nil {
				return fmt.Errorf("写入音频包失败: %w", err)
			}
			packetCount++
		}
	}

	if !hasAudio {
		return fmt.Errorf("视频中未找到音频轨道")
	}

	if err := muxer.WriteTrailer(); err != nil {
		return fmt.Errorf("写入 m4a 尾部失败: %w", err)
	}

	utils.Info("纯 Go 抽轨完成，共 %d 个音频包", packetCount)
	return nil
}
