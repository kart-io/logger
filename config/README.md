# Config Package

配置管理包，提供智能的多源配置解析和冲突处理，特别是针对 OTLP 配置的智能检测和自动启用功能。

## 概述

`config` 包实现了项目需求文档中定义的核心配置管理功能：

- **多源配置支持**：环境变量、YAML文件、JSON格式
- **智能 OTLP 配置**：基于端点自动启用，消除冗余配置
- **扁平化配置**：顶级字段用于常见场景，嵌套配置用于高级控制
- **配置冲突解决**：基于优先级的智能处理
- **配置验证**：完整性检查和默认值填充

## 核心类型

### Config 结构体

```go
type Config struct {
    Engine       string   `yaml:"engine" json:"engine" env:"LOG_ENGINE"`
    Level        string   `yaml:"level" json:"level" env:"LOG_LEVEL"`
    Format       string   `yaml:"format" json:"format" env:"LOG_FORMAT"`
    OutputPaths  []string `yaml:"output-paths" json:"output_paths" env:"LOG_OUTPUT_PATHS"`
    
    // 扁平化 OTLP 配置（常用场景）
    OTLPEndpoint string      `yaml:"otlp-endpoint" json:"otlp_endpoint" env:"LOG_OTLP_ENDPOINT"`
    
    // 嵌套 OTLP 配置（高级控制）
    OTLP         *OTLPConfig `yaml:"otlp" json:"otlp"`
    
    Development       bool `yaml:"development" json:"development" env:"LOG_DEVELOPMENT"`
    DisableCaller     bool `yaml:"disable-caller" json:"disable_caller" env:"LOG_DISABLE_CALLER"`
    DisableStacktrace bool `yaml:"disable-stacktrace" json:"disable_stacktrace" env:"LOG_DISABLE_STACKTRACE"`
}
```

### OTLPConfig 结构体

```go
type OTLPConfig struct {
    Enabled  *bool             `yaml:"enabled" json:"enabled" env:"LOG_OTLP_ENABLED"`
    Endpoint string            `yaml:"endpoint" json:"endpoint" env:"LOG_OTLP_ENDPOINT"`
    Protocol string            `yaml:"protocol" json:"protocol" env:"LOG_OTLP_PROTOCOL"`
    Timeout  time.Duration     `yaml:"timeout" json:"timeout" env:"LOG_OTLP_TIMEOUT"`
    Headers  map[string]string `yaml:"headers" json:"headers"`
}
```

## 智能 OTLP 配置

### 设计原则

根据需求文档，OTLP 配置遵循以下智能原则：

1. **智能检测**：有端点配置自动启用 OTLP
2. **用户意图优先**：明确的 `enabled: false` 总是被尊重
3. **扁平化优先**：顶级 `otlp-endpoint` 用于简单场景
4. **嵌套控制**：`otlp.*` 配置用于复杂场景

### 配置优先级

1. 环境变量 (最高优先级)
2. 显式的 `enabled` 设置
3. 端点存在时的自动启用
4. 默认值

### 使用示例

#### 1. 最简配置（自动启用）
```yaml
# logger.yaml
otlp-endpoint: "http://localhost:4317"
# OTLP 将自动启用
```

#### 2. 完整扁平化配置
```yaml
# logger.yaml  
engine: "zap"
level: "debug"
format: "json"
otlp-endpoint: "http://jaeger:4317"
development: true
```

#### 3. 嵌套配置（高级控制）
```yaml
# logger.yaml
engine: "slog"
level: "info"
otlp:
  enabled: true
  endpoint: "https://otlp.example.com:4317"
  protocol: "grpc"
  timeout: "15s"
  headers:
    Authorization: "Bearer token123"
    X-Tenant-ID: "tenant-456"
```

#### 4. 环境变量配置
```bash
export LOG_ENGINE="zap"
export LOG_LEVEL="info"
export LOG_OTLP_ENDPOINT="http://localhost:4317"
export LOG_OTLP_PROTOCOL="http"
```

#### 5. 显式禁用 OTLP
```yaml
# logger.yaml
otlp-endpoint: "http://localhost:4317"  # 有端点
otlp:
  enabled: false  # 但明确禁用，优先级更高
```

## 配置冲突处理

### 扁平化 vs 嵌套配置

当同时存在扁平化和嵌套配置时：

```yaml
# 配置示例
otlp-endpoint: "http://simple.example.com:4317"  # 扁平化
otlp:
  endpoint: "http://advanced.example.com:4317"    # 嵌套
  enabled: false                                   # 嵌套
```

**解决逻辑**：
1. 如果 `otlp.enabled: false`，尊重用户意图，禁用 OTLP
2. 如果 `otlp.endpoint` 为空，使用扁平化的 `otlp-endpoint`
3. 如果 `otlp.endpoint` 有值，使用嵌套配置

### 环境变量覆盖

```bash
# 文件配置
otlp-endpoint: "http://localhost:4317"

# 环境变量（优先级更高）
export LOG_OTLP_ENDPOINT="http://production.example.com:4317"
```

结果：使用环境变量值 `http://production.example.com:4317`

## API 使用

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/kart-io/logger/config"
)

func main() {
    // 创建默认配置
    cfg := config.DefaultConfig()
    
    // 设置 OTLP 端点（自动启用）
    cfg.OTLPEndpoint = "http://localhost:4317"
    
    // 验证和解析配置
    if err := cfg.Validate(); err != nil {
        panic(err)
    }
    
    fmt.Println("OTLP enabled:", cfg.IsOTLPEnabled())
}
```

### 从文件加载配置

```go
package main

import (
    "github.com/kart-io/logger/config"
    "gopkg.in/yaml.v3"
    "os"
)

func loadConfigFromFile(path string) (*config.Config, error) {
    // 从默认配置开始
    cfg := config.DefaultConfig()
    
    // 读取文件
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    // 解析 YAML
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }
    
    // 验证配置
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return cfg, nil
}
```

### 环境变量集成

```go
package main

import (
    "github.com/kart-io/logger/config"
    "os"
)

func configWithEnv() *config.Config {
    cfg := config.DefaultConfig()
    
    // 手动处理环境变量（或使用 envconfig 等库）
    if engine := os.Getenv("LOG_ENGINE"); engine != "" {
        cfg.Engine = engine
    }
    
    if level := os.Getenv("LOG_LEVEL"); level != "" {
        cfg.Level = level
    }
    
    if endpoint := os.Getenv("LOG_OTLP_ENDPOINT"); endpoint != "" {
        cfg.OTLPEndpoint = endpoint
    }
    
    cfg.Validate()
    return cfg
}
```

## 默认值

```go
func DefaultConfig() *Config {
    return &Config{
        Engine:            "slog",           // 使用标准库 slog
        Level:             "INFO",           // 信息级别
        Format:            "json",           // JSON 格式输出
        OutputPaths:       []string{"stdout"}, // 标准输出
        Development:       false,            // 生产模式
        DisableCaller:     false,            // 启用调用者信息
        DisableStacktrace: false,            // 启用堆栈跟踪
        OTLP: &OTLPConfig{
            Protocol: "grpc",                // gRPC 协议
            Timeout:  10 * time.Second,      // 10秒超时
        },
    }
}
```

## 验证规则

`Validate()` 方法执行以下检查：

1. **日志级别验证**：确保级别字符串可以解析
2. **引擎验证**：只支持 "zap" 和 "slog"
3. **OTLP 智能解析**：根据端点和显式设置决定启用状态
4. **默认值填充**：为未设置的必要字段提供默认值

## 配置状态术语

- **禁用 (disabled)**：用户明确设置 `enabled: false`
- **未启用 (not enabled)**：没有端点配置，无法启用
- **自动禁用 (auto disabled)**：系统智能判断应禁用

## 扩展性

配置包设计为可扩展：

```go
// 自定义配置结构
type CustomConfig struct {
    *config.Config
    CustomField string `yaml:"custom_field"`
}

// 自定义验证逻辑
func (c *CustomConfig) CustomValidate() error {
    if err := c.Config.Validate(); err != nil {
        return err
    }
    
    // 自定义验证逻辑
    if c.CustomField == "" {
        c.CustomField = "default_value"
    }
    
    return nil
}
```

## 注意事项

1. **环境变量优先级**：环境变量在启动时读取，运行时变更需要重载机制
2. **OTLP 自动启用**：有端点配置会自动启用，除非明确设置 `enabled: false`
3. **配置验证**：总是调用 `Validate()` 确保配置的一致性和完整性
4. **指针字段**：`OTLP.Enabled` 使用指针以区分未设置、true 和 false 三种状态
5. **时间格式**：`Timeout` 字段支持 Go duration 格式（如 "10s", "1m30s"）

## 相关包

- [`option`](../option/) - 配置选项和参数处理
- [`factory`](../factory/) - 使用配置创建日志器
- [`otlp`](../otlp/) - OTLP 功能实现