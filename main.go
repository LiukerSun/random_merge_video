package main

import (
	"embed"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

//go:embed ffmpeg/*
var ffmpegFiles embed.FS

type VideoInfo struct {
	Path     string
	Duration float64
}

type Config struct {
	NumCombinations int
	TargetDuration  float64
	MinDuration     float64
	MaxVideos       int
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func getVideoDuration(videoPath string) (float64, error) {
	ffprobePath := "ffmpeg/ffprobe.exe"
	cmd := exec.Command(ffprobePath, "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("获取视频时长失败: %v", err)
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("解析视频时长失败: %v", err)
	}
	return duration, nil
}

func cutVideo(inputPath, outputPath string, startTime, duration float64) error {
	ffmpegPath := "ffmpeg/ffmpeg.exe"
	cmd := exec.Command(ffmpegPath, "-y", "-i", inputPath, "-ss", fmt.Sprintf("%.2f", startTime),
		"-t", fmt.Sprintf("%.2f", duration), "-c:v", "libx264", "-an", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("裁剪视频失败: %v\n输出: %s", err, string(output))
	}
	return nil
}

func loadConfig() (*Config, error) {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		return nil, fmt.Errorf("读取配置文件错误: %v", err)
	}

	config := &Config{}
	config.NumCombinations = cfg.Section("").Key("num_combinations").MustInt(5)
	config.TargetDuration = cfg.Section("").Key("target_duration").MustFloat64(60.0)
	config.MinDuration = cfg.Section("").Key("min_duration").MustFloat64(5.0)
	config.MaxVideos = cfg.Section("").Key("max_videos").MustInt(10)

	return config, nil
}

func main() {
	// 检查并解压ffmpeg文件
	if err := extractFFmpeg(); err != nil {
		fmt.Printf("解压ffmpeg文件失败: %v\n", err)
		return
	}
	defer os.RemoveAll("ffmpeg")

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 加载配置
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 检查source_videos目录
	sourceDir := "source_videos"
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		fmt.Println("错误：source_videos目录不存在")
		return
	}

	// 获取所有视频文件
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		fmt.Printf("读取目录错误: %v\n", err)
		return
	}

	var videoInfos []VideoInfo
	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".mkv" {
				videoPath := filepath.Join(sourceDir, file.Name())
				duration, err := getVideoDuration(videoPath)
				if err != nil {
					fmt.Printf("获取视频时长错误 %s: %v\n", videoPath, err)
					continue
				}
				if duration < config.MinDuration {
					fmt.Printf("跳过视频: %s, 因为时长 %.2f秒小于最小要求 %.2f秒\n", file.Name(), duration, config.MinDuration)
					continue
				}
				fmt.Printf("找到视频: %s, 时长: %.2f秒\n", file.Name(), duration)
				videoInfos = append(videoInfos, VideoInfo{
					Path:     videoPath,
					Duration: duration,
				})
			}
		}
	}

	// 如果视频数量过多，随机选择一些视频
	if len(videoInfos) > config.MaxVideos {
		fmt.Printf("视频数量过多，随机选择%d个视频\n", config.MaxVideos)
		videoInfos = videoInfos[:config.MaxVideos]
		rand.Shuffle(len(videoInfos), func(i, j int) {
			videoInfos[i], videoInfos[j] = videoInfos[j], videoInfos[i]
		})
	}

	if len(videoInfos) < 2 {
		fmt.Println("错误：需要至少2个有效视频文件")
		return
	}

	// 计算最大可能的组合数量
	maxCombinations := factorial(len(videoInfos))
	if config.NumCombinations > maxCombinations {
		fmt.Printf("警告：请求的组合数量(%d)超过了最大可能组合数(%d)，将使用最大可能组合数\n",
			config.NumCombinations, maxCombinations)
		config.NumCombinations = maxCombinations
	}

	// 创建results目录
	if err := os.MkdirAll("results", 0755); err != nil {
		fmt.Printf("创建results目录错误: %v\n", err)
		return
	}

	// 生成指定数量的不同随机组合
	for i := 0; i < config.NumCombinations; i++ {
		fmt.Printf("\n正在生成第 %d 个组合视频...\n", i+1)

		// 打乱视频顺序
		shuffled := make([]VideoInfo, len(videoInfos))
		copy(shuffled, videoInfos)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		// 创建临时文件列表
		tempFile, err := os.CreateTemp("", "filelist-*.txt")
		if err != nil {
			fmt.Printf("创建临时文件错误: %v\n", err)
			continue
		}
		defer os.Remove(tempFile.Name())

		// 计算每个视频需要裁剪的时长
		remainingDuration := config.TargetDuration
		var tempClips []string
		var totalDuration float64

		// 计算所有视频的总时长
		totalAvailableDuration := 0.0
		for _, video := range shuffled {
			totalAvailableDuration += video.Duration - 1 // 每个视频至少保留1秒
		}

		// 如果总可用时长小于目标时长，跳过这个组合
		if totalAvailableDuration < config.TargetDuration {
			fmt.Printf("警告：可用视频总时长(%.2f秒)小于目标时长(%.2f秒)，跳过此组合\n",
				totalAvailableDuration, config.TargetDuration)
			continue
		}

		// 按比例分配每个视频的时长
		for _, video := range shuffled {
			if remainingDuration <= 0 {
				break
			}

			// 计算这个视频应该贡献的时长比例
			videoAvailableDuration := video.Duration - 1
			videoRatio := videoAvailableDuration / totalAvailableDuration
			clipDuration := config.TargetDuration * videoRatio

			// 确保每个视频片段至少有minDuration
			if clipDuration < config.MinDuration {
				clipDuration = config.MinDuration
			}

			// 如果这是最后一个视频，使用剩余时长
			if remainingDuration-clipDuration < config.MinDuration {
				clipDuration = remainingDuration
			}

			// 随机选择裁剪的起始时间
			maxStartTime := video.Duration - clipDuration - 1
			if maxStartTime <= 0 {
				continue
			}
			startTime := rand.Float64() * maxStartTime

			// 创建裁剪后的临时视频文件
			tempClip, err := os.CreateTemp("", "clip-*.mp4")
			if err != nil {
				fmt.Printf("创建临时剪辑文件错误: %v\n", err)
				continue
			}
			tempClip.Close()
			tempClips = append(tempClips, tempClip.Name())
			defer os.Remove(tempClip.Name())

			fmt.Printf("裁剪视频: %s, 起始时间: %.2f秒, 时长: %.2f秒\n",
				filepath.Base(video.Path), startTime, clipDuration)

			// 裁剪视频
			if err := cutVideo(video.Path, tempClip.Name(), startTime, clipDuration); err != nil {
				fmt.Printf("裁剪视频错误: %v\n", err)
				continue
			}

			// 写入文件列表
			absPath, err := filepath.Abs(tempClip.Name())
			if err != nil {
				fmt.Printf("获取绝对路径错误: %v\n", err)
				continue
			}
			absPath = strings.ReplaceAll(absPath, "\\", "\\\\")
			fmt.Fprintf(tempFile, "file '%s'\n", absPath)

			totalDuration += clipDuration
			remainingDuration -= clipDuration
		}

		// 检查总时长是否达到目标
		if totalDuration < config.TargetDuration {
			fmt.Printf("警告：无法生成足够时长的视频，实际时长: %.2f秒，目标时长: %.2f秒\n",
				totalDuration, config.TargetDuration)
			continue
		}

		tempFile.Close()

		// 生成输出文件名
		outputFile := fmt.Sprintf("results/combined_%d.mp4", i+1)

		// 使用ffmpeg合并视频
		cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", tempFile.Name(),
			"-c:v", "libx264", "-an", outputFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("生成视频错误: %v\n输出: %s\n", err, string(output))
			continue
		}
		fmt.Printf("成功生成: %s, 总时长: %.2f秒\n", outputFile, totalDuration)
	}
}

func extractFFmpeg() error {
	// 创建ffmpeg目录
	if err := os.MkdirAll("ffmpeg", 0755); err != nil {
		return err
	}

	// 解压ffmpeg文件
	files, err := ffmpegFiles.ReadDir("ffmpeg")
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := ffmpegFiles.ReadFile("ffmpeg/" + file.Name())
		if err != nil {
			return err
		}

		if err := os.WriteFile("ffmpeg/"+file.Name(), data, 0755); err != nil {
			return err
		}
	}

	return nil
}
