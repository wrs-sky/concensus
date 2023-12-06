package benchmark

import (
	"fmt"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Configuration 结构体用于映射 YAML 文件的结构
type Configuration struct {
	Server  ServerConfiguration `yaml:"Server"`
	Block   BlockConfiguration  `yaml:"Block"`
	Log     LogConfiguration    `yaml:"Log"`
	System  SystemConfiguration `yaml:"System"`
	WorkDir string
	Suffix  string
}

type ServerConfiguration struct {
	Num       int `yaml:"Num"`
	BatchSize int `yaml:"BatchSize"`
}

type BlockConfiguration struct {
	Count int `yaml:"Count"`
}

type LogConfiguration struct {
	LogDir  string `yaml:"LogDir"`
	TestDir string `yaml:"TestDir"`
}

type SystemConfiguration struct {
	Timeout int `yaml:"Timeout"`
}

// InitConfig 用于初始化加载配置文件
func InitConfig(workDir string, confFile string) (*Configuration, error) {
	data, err := ioutil.ReadFile(filepath.Join(workDir, confFile))
	if err != nil {
		fmt.Errorf("ioutil.ReadFile error:%v", err)
		return nil, err
	}

	c := &Configuration{}
	if err = yaml.Unmarshal(data, &c); err != nil {
		fmt.Errorf("yaml.Unmarshal error:%v", err)
		return nil, err
	}

	c.WorkDir = workDir

	// 指定文件后缀
	fileSuffix := time.Now().Format("20060102150405")
	c.Suffix = fileSuffix
	fmt.Println("Configuration suffix:", c.Suffix)

	//重置路径
	c.Log.LogDir = filepath.Join(c.WorkDir, c.Log.LogDir+c.Suffix)
	c.Log.TestDir = filepath.Join(c.WorkDir, c.Log.TestDir+c.Suffix)

	if err := c.Setup(); err != nil {
		return c, err
	}

	fmt.Println("Configuration setup successfully")
	return c, nil
}

// Setup 用于初始化配置
func (c *Configuration) Setup() error {

	//日志文件夹生成
	logDir := c.Log.LogDir
	perm := os.FileMode(0777)
	if err := os.Mkdir(logDir, perm); err != nil {
		fmt.Println("Error creating logDir directory:", err)
		return err
	}

	//tmp文件夹生成
	testDir := c.Log.TestDir
	if err := os.Mkdir(testDir, perm); err != nil {
		fmt.Println("Error creating testDir directory:", err)
		return err
	}

	return nil
}
func NewLogger(logFilePath string) (smart.Logger, error) {
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()

	// 配置日志记录器
	logConfig := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{logFilePath},
		ErrorOutputPaths: []string{logFilePath},
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}
	basicLog, err := logConfig.Build()
	if err != nil {
		fmt.Println("Failed to initialize logger:", err)
		return nil, err
	}

	return basicLog.Sugar(), nil
}
