package main

import (
	"context"
	"exam-system/internal/config"
	"exam-system/internal/middleware"
	"exam-system/internal/pkg/banner"
	"exam-system/internal/pkg/database"
	"exam-system/internal/pkg/logger"
	"exam-system/internal/router"
	"exam-system/internal/service"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v3"
)

//go:generate echo "嵌入静态资源..."

// 版本信息，编译时通过 ldflags 设置
var (
	Version    = "v0.1.2"
	CommitHash = "unknown"
	BuildTime  = "unknown"
)

func main() {
	// 创建 CLI 应用
	cmd := &cli.Command{
		Name:    "SimpleExam API Server",
		Usage:   "在线考试系统后端服务",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "配置文件路径",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			configPath, err := resolveConfigPath(cmd)
			if err != nil {
				return err
			}

			// 将配置文件路径设置到环境变量中，供config包读取
			os.Setenv("CONFIG_PATH", configPath)

			// 启动应用
			return startApp()
		},
		Commands: []*cli.Command{
			{
				Name:    "reset-password",
				Aliases: []string{"r"},
				Usage:   "重置指定用户的密码",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Aliases:  []string{"u"},
						Usage:    "需要重置密码的用户名",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "新的密码",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "配置文件路径",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath, err := resolveConfigPath(cmd)
					if err != nil {
						return err
					}
					os.Setenv("CONFIG_PATH", configPath)

					// 初始化配置和依赖
					if _, err := config.Load(); err != nil {
						return fmt.Errorf("加载配置失败: %v", err)
					}
					if err := logger.Setup(); err != nil {
						return fmt.Errorf("初始化日志系统失败: %v", err)
					}
					if err := database.Setup(); err != nil {
						return fmt.Errorf("数据库初始化失败: %v", err)
					}

					username := cmd.String("username")
					password := cmd.String("password")

					if err := service.Auth.ResetPassword(username, password); err != nil {
						return fmt.Errorf("重置密码失败: %v", err)
					}

					fmt.Printf("用户 %s 的密码已重置。\n", username)
					return nil
				},
			},
		},
	}

	// 运行应用
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("应用程序启动失败: %v", err)
	}
}

// startApp 启动应用程序的主要逻辑
func startApp() error {
	// 打印启动横幅
	banner.Print(Version, CommitHash, BuildTime)

	// 加载配置
	_, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 初始化日志系统
	err = logger.Setup()
	if err != nil {
		return fmt.Errorf("初始化日志系统失败: %v", err)
	}

	logger.Info("配置加载完成")

	// 初始化数据库
	err = database.Setup()
	if err != nil {
		logger.Fatalf("数据库初始化失败: %v", err)
		return fmt.Errorf("数据库初始化失败: %v", err)
	}

	// 确保默认管理员账号存在
	if err := service.Auth.EnsureDefaultAdmin(); err != nil {
		logger.Fatalf("初始化管理员账号失败: %v", err)
		return fmt.Errorf("初始化管理员账号失败: %v", err)
	}

	logger.Info("数据库初始化完成")

	// 启动定时任务
	service.Cron.Start()
	defer service.Cron.Stop()
	logger.Info("定时任务启动完成")

	// 设置gin模式
	gin.SetMode(config.GlobalConfig.Server.Mode)
	if config.GlobalConfig.Server.Mode == "release" {
		logger.Info("Gin设置为生产模式")
	} else {
		logger.Info("Gin运行在调试模式")
	}

	// 初始化路由（不使用默认中间件）
	r := gin.New()
	// 添加我们自定义的中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// 获取静态文件系统
	userFS := GetUserFS()
	adminFS := GetAdminFS()
	logger.Info("静态文件系统初始化完成")

	// 使用统一的路由设置函数
	router.SetupRoutes(r, userFS, adminFS)
	logger.Info("路由设置完成")

	// 启动服务器
	logger.Infof("服务器启动中，端口: %s", config.GlobalConfig.Server.Port)
	err = r.Run(":" + config.GlobalConfig.Server.Port)
	if err != nil {
		logger.Fatalf("服务器启动失败: %v", err)
		return fmt.Errorf("服务器启动失败: %v", err)
	}

	return nil
}

// resolveConfigPath 获取配置文件路径
func resolveConfigPath(cmd *cli.Command) (string, error) {
	configPath := cmd.String("config")

	// 如果未指定配置文件，尝试从默认位置加载
	if configPath == "" {
		possiblePaths := []string{
			"config.yaml",
			filepath.Join("config", "config.yaml"),
		}

		found := false
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("未指定配置文件且未找到默认配置文件(config.yaml或config/config.yaml)")
		}
	}

	return configPath, nil
}
