package main

import (
	"fmt"
	"github.com/adrg/xdg"
)

func main() {
	// 程序配置目录 ~/.config/你的应用名
	fmt.Println("配置根目录:", xdg.ConfigHome)
	// 程序数据存储目录 ~/.local/share/你的应用名
	fmt.Println("数据根目录:", xdg.DataHome)
	// 缓存目录
	fmt.Println("缓存目录:", xdg.CacheHome)

	// 系统用户常用目录
	fmt.Println("下载目录:", xdg.UserDirs.Download)
	fmt.Println("文档目录:", xdg.UserDirs.Documents)
	fmt.Println("图片目录:", xdg.UserDirs.Pictures)
	fmt.Println("视频目录:", xdg.UserDirs.Videos)
	fmt.Println("桌面目录:", xdg.UserDirs.Desktop)
}

