# 命令行小工具`play-all-mp3`：播放目录中的所有`MP3`

功能概括：播放目录中的所有MP3.。

播放顺序：按照文件名排序。

继续播放：关闭程序时自动记录播放进度，下次播放时自动继续播放。

用法：

```
格式：play-all-mp3 [选项] <保存着MP3文件的目录>
选项:
  -fc
    	从上次播放的文件头部开始播放
  -h	显示帮助信息。
  -nc
    	从头播放，不读取播放进度
```

源代码：[https://gitee.com/rocket049/play-all-mp3](https://gitee.com/rocket049/play-all-mp3)

编译安装：

1. 安装`go`编译环境：[go编译器](https://golang.google.cn/)
2. 安装`libasound2-dev`
3. 编译命令：`go get -u -v gitee.com/rocket049/play-all-mp3`

现在`play-all-mp3`程序已经被安装到`GOPATH/bin`目录中。
