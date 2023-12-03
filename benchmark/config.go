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

var workDir string

// Configuration 结构体用于映射 YAML 文件的结构
type Configuration struct {
	Server  ServerConfiguration `yaml:"Server"`
	Block   BlockConfiguration  `yaml:"Block"`
	Log     Log                 `yaml:"Log"`
	WorkDir string
	Suffix  string
	Logger  smart.Logger
}

type ServerConfiguration struct {
	Num int `yaml:"Num"`
}
type BlockConfiguration struct {
	Count int `yaml:"Count"`
}
type Log struct {
	LogFile string `yaml:"LogFile"`
	TestDir string `yaml:"TestDir"`
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

	//重置路径
	c.Log.LogFile = filepath.Join(c.WorkDir, c.Log.LogFile+c.Suffix+".log")
	c.Log.TestDir = filepath.Join(c.WorkDir, c.Log.TestDir+c.Suffix)

	if err := c.Setup(); err != nil {
		return c, err
	}

	return c, nil
}

// Setup 用于初始化配置
func (c *Configuration) Setup() error {

	// 指定输出日志文件路径
	var logFilePath = c.Log.LogFile

	// 创建日志文件
	logFile, err := os.Create(logFilePath)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	time.Sleep(1 * time.Second)

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
		return err
	}

	c.Logger = basicLog.Sugar()
	fmt.Println("Log configuration succeeded")

	//tmp文件生成
	testDir = c.Log.TestDir
	perm := os.FileMode(0777)
	if err := os.Mkdir(testDir, perm); err != nil {
		fmt.Println("Error creating directory:", err)
		return err
	}

	return nil
}
