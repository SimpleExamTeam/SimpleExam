package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`

	Database struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`

	JWT struct {
		Secret     string `yaml:"secret"`
		ExpireTime int    `yaml:"expire_time"`
	} `yaml:"jwt"`

	Log struct {
		Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error
		Format     string `yaml:"format"`      // 日志格式: json, text
		Output     string `yaml:"output"`      // 输出方式: console, file, both
		FilePath   string `yaml:"file_path"`   // 日志文件路径
		MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小(MB)
		MaxBackups int    `yaml:"max_backups"` // 保留的旧日志文件数量
		MaxAge     int    `yaml:"max_age"`     // 日志文件保留天数
		Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
	} `yaml:"log"`

	WeChat struct {
		AppID               string `yaml:"app_id"`
		AppSecret           string `yaml:"app_secret"`
		MchID               string `yaml:"mch_id"`
		PayKey              string `yaml:"pay_key"`
		NotifyURL           string `yaml:"notify_url"`
		OAuthRedirect       string `yaml:"oauth_redirect"`
		AdminOAuthRedirect  string `yaml:"admin_oauth_redirect"` // 管理员微信网页授权回调地址
		QRCodeCallback      string `yaml:"qrcode_callback"`
		AdminQRCodeCallback string `yaml:"admin_qrcode_callback"` // 管理员二维码扫码回调地址
		RefundURL           string `yaml:"refund_url"`            // 退款接口URL
		RefundNotifyURL     string `yaml:"refund_notify_url"`     // 退款回调通知URL
		RefundQueryURL      string `yaml:"refund_query_url"`      // 退款查询URL
		CertPath            string `yaml:"cert_path"`             // 商户证书路径
	} `yaml:"wechat"`
}

var GlobalConfig *Config

func Load() (*Config, error) {
	if GlobalConfig != nil {
		return GlobalConfig, nil
	}

	// 获取配置文件路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// 如果环境变量中没有配置路径，则使用默认路径
		// 获取当前工作目录
		workDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("获取工作目录失败: %v", err)
		}

		// 尝试默认配置路径
		configPath = filepath.Join(workDir, "config", "config.yaml")

		// 如果默认配置不存在，尝试根目录下的配置文件
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = filepath.Join(workDir, "config.yaml")
		}
	}

	// 读取配置文件
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败 %s: %v", configPath, err)
	}

	// 解析配置文件
	config := &Config{}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 添加调试信息
	fmt.Printf("加载配置文件: %s\n", configPath)
	fmt.Printf("微信配置: %+v\n", config.WeChat)

	// 设置默认值
	// 日志配置默认值
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Format == "" {
		config.Log.Format = "text"
	}
	if config.Log.Output == "" {
		config.Log.Output = "console"
	}
	if config.Log.FilePath == "" {
		config.Log.FilePath = "logs/app.log"
	}
	if config.Log.MaxSize == 0 {
		config.Log.MaxSize = 100 // 100MB
	}
	if config.Log.MaxBackups == 0 {
		config.Log.MaxBackups = 3
	}
	if config.Log.MaxAge == 0 {
		config.Log.MaxAge = 28 // 28天
	}

	// 微信配置默认值
	if config.WeChat.RefundURL == "" {
		config.WeChat.RefundURL = "https://api.mch.weixin.qq.com/secapi/pay/refund"
	}
	if config.WeChat.RefundQueryURL == "" {
		config.WeChat.RefundQueryURL = "https://api.mch.weixin.qq.com/pay/refundquery"
	}
	if config.WeChat.CertPath == "" {
		config.WeChat.CertPath = "certs/apiclient_cert.p12"
	}

	GlobalConfig = config
	return config, nil
}
