# 随机视频合并工具

这是一个用于随机合并视频片段的工具，可以生成多个不同组合的视频。

## 功能特点

- 随机选择视频片段进行组合
- 支持设置目标视频时长
- 支持设置每个片段的最小时长
- 自动去除音频，只保留画面
- 无需安装 ffmpeg，程序自带

## 使用前准备

1. 下载 ffmpeg 文件：
   - 从 [ffmpeg 官网](https://ffmpeg.org/download.html) 下载 Windows 版本
   - 将 `ffmpeg.exe` 和 `ffprobe.exe` 放入 `ffmpeg` 目录

2. 准备视频素材：
   - 在程序目录下创建 `source_videos` 文件夹
   - 将需要合并的视频文件放入该文件夹
   - 支持的视频格式：mp4, avi, mov, mkv

## 配置说明

编辑 `config.ini` 文件调整参数：

```ini
; 要生成的视频组合数量
num_combinations = 5

; 目标视频时长（秒）
target_duration = 60.0

; 每个视频的最小时长（秒）
min_duration = 5.0

; 最大使用的视频数量
max_videos = 10
```

参数说明：
- `num_combinations`: 要生成的视频组合数量
- `target_duration`: 目标视频时长（秒）
- `min_duration`: 每个视频片段的最小时长（秒）
- `max_videos`: 最大使用的视频数量

## 使用方法

1. 编译程序：
   ```bash
   build.bat
   ```

2. 运行程序：
   ```bash
   video_combiner.exe
   ```

3. 查看结果：
   - 生成的视频将保存在 `results` 目录中
   - 文件名格式：`combined_1.mp4`, `combined_2.mp4` 等

## 注意事项

1. 视频素材要求：
   - 每个视频时长应至少为 `min_duration + 1` 秒
   - 建议准备多个视频素材，以获得更好的随机效果

2. 如果出现以下情况，程序会跳过当前组合：
   - 可用视频总时长小于目标时长
   - 无法生成足够时长的视频片段

3. 程序运行时会显示详细日志，包括：
   - 找到的视频文件及其时长
   - 每个视频片段的裁剪信息
   - 生成结果的状态

## 常见问题

1. 程序无法启动：
   - 检查是否已放入 ffmpeg 文件
   - 检查 `source_videos` 目录是否存在

2. 生成的视频时长不足：
   - 增加 `min_duration` 的值
   - 准备更多或更长的视频素材

3. 生成的视频数量少于预期：
   - 检查视频素材数量是否足够
   - 调整 `max_videos` 参数 