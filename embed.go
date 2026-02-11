package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed public/user
var userFS embed.FS

//go:embed public/admin
var adminFS embed.FS

// GetUserFS 获取用户端静态文件系统
func GetUserFS() http.FileSystem {
	sub, err := fs.Sub(userFS, "public/user")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// GetAdminFS 获取管理端静态文件系统
func GetAdminFS() http.FileSystem {
	sub, err := fs.Sub(adminFS, "public/admin")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// 调试函数：打印文件系统中的文件
func printFSFiles(name string, fsys fs.FS) {
	fmt.Printf("\n--- %s 文件列表 ---\n", name)
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("访问 %s 错误: %v\n", path, err)
			return nil
		}
		if !d.IsDir() {
			fmt.Printf("- %s\n", path)
		}
		return nil
	})
}
