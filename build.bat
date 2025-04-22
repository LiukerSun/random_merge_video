@echo off
chcp 65001 > nul
echo 正在检查ffmpeg文件...
if not exist ffmpeg\ffmpeg.exe (
    echo 错误：ffmpeg.exe文件不存在！
    echo 请从ffmpeg官网下载Windows版本的ffmpeg，并将以下文件放入ffmpeg目录中：
    echo - ffmpeg.exe
    echo - ffprobe.exe
    pause
    exit /b 1
)

echo 正在编译程序...
go build -o video_combiner.exe main.go
echo 编译完成！
echo.
echo 使用方法：
echo 1. 将视频文件放入source_videos目录
echo 2. 运行video_combiner.exe
echo 3. 生成的视频将保存在results目录中
echo.
pause 