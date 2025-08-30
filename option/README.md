# Option Package

配置选项包，提供完整的日志库配置管理、验证和智能默认值。支持多层级配置冲突处理和 OTLP 自动检测。

## 📋 特性

- ✅ **完整配置结构**: 涵盖所有日志引擎和 OTLP 配置项
- ✅ **智能默认值**: 开箱即用的合理配置
- ✅ **配置验证**: 自动验证配置一致性和有效性
- ✅ **命令行标志**: 完整的 pflag 集成支持
- ✅ **OTLP 智能配置**: 基于端点自动启用/禁用 OTLP
- ✅ **配置冲突处理**: 支持扁平化和嵌套配置优先级
- ✅ **类型安全**: 完整的结构体标签支持 JSON 和 mapstructure

## 🚀 快速使用

### 基础配置

```go
package main

import (
    "fmt"
    "github.com/kart-io/logger/option"
)

func main() {
    // 使用默认配置
    opt := option.DefaultLogOption()
    
    fmt.Printf("引擎: %s\n", opt.Engine)        // slog
    fmt.Printf("级别: %s\n", opt.Level)         // INFO
    fmt.Printf("格式: %s\n", opt.Format)        // json
    fmt.Printf("输出: %v\n", opt.OutputPaths)   // [stdout]
}
```

### 自定义配置

```go
// 创建自定义配置
opt := &option.LogOption{
    Engine:      "zap",           // 使用高性能 Zap 引擎
    Level:       "DEBUG",         // 设置调试级别
    Format:      "console",       // 控制台友好格式
    OutputPaths: []string{"stdout", "/var/log/app.log"}, // 多输出
    Development: true,            // 开发模式
    
    // OTLP 配置
    OTLPEndpoint: "http://localhost:4317", // 扁平化配置
    OTLP: &option.OTLPOption{
        Protocol: "grpc",
        Timeout:  15 * time.Second,
        Headers: map[string]string{
            "Authorization": "Bearer token123",
        },
    },
}

// 验证配置
if err := opt.Validate(); err != nil {
    panic(err)
}
```

## 🔧 配置结构

### LogOption 主配置

```go
type LogOption struct {
    // 核心引擎配置
    Engine string `json:"engine"`                    // "zap" 或 "slog"
    Level  string `json:"level"`                     // 日志级别
    Format string `json:"format"`                    // 输出格式
    
    // 输出配置
    OutputPaths []string `json:"output_paths"`       // 输出目标
    
    // OTLP 配置（扁平化和嵌套）
    OTLPEndpoint string      `json:"otlp_endpoint"`  // 扁平化端点
    OTLP         *OTLPOption `json:"otlp"`           // 嵌套配置
    
    // 功能开关
    Development       bool `json:"development"`        // 开发模式
    DisableCaller     bool `json:"disable_caller"`     // 禁用调用者
    DisableStacktrace bool `json:"disable_stacktrace"` // 禁用堆栈
}
```

### OTLPOption OTLP配置

```go
type OTLPOption struct {
    Enabled  *bool             `json:"enabled"`   // 启用状态（三态逻辑）
    Endpoint string            `json:"endpoint"`  // OTLP 端点
    Protocol string            `json:"protocol"`  // 协议类型
    Timeout  time.Duration     `json:"timeout"`   // 超时时间
    Headers  map[string]string `json:"headers"`   // 请求头
    Insecure bool              `json:"insecure"`  // 不安全连接
}
```

## ⚙️ 配置方式

### 1. 代码配置

```go
// 高性能生产配置
opt := &option.LogOption{
    Engine:      "zap",
    Level:       "INFO",
    Format:      "json",
    OutputPaths: []string{"/var/log/app.log"},
    Development: false,
    OTLPEndpoint: "https://otlp.company.com:4317",
}

// 开发调试配置
devOpt := &option.LogOption{
    Engine:      "slog",
    Level:       "DEBUG", 
    Format:      "console",
    OutputPaths: []string{"stdout"},
    Development: true,
}
```

### 2. 命令行标志

```go
import "github.com/spf13/pflag"

func main() {
    opt := option.DefaultLogOption()
    
    // 添加到 pflag.FlagSet
    fs := pflag.NewFlagSet("logger", pflag.ExitOnError)
    opt.AddFlags(fs)
    
    // 解析命令行参数
    fs.Parse(os.Args[1:])
    
    // 验证配置
    if err := opt.Validate(); err != nil {
        log.Fatal(err)
    }
}
```

使用示例：
```bash
./app --engine=zap --level=DEBUG --format=console --otlp-endpoint=http://localhost:4317
```

### 3. JSON 配置文件

```json
{
  "engine": "zap",
  "level": "INFO",
  "format": "json",
  "output_paths": ["stdout", "/var/log/app.log"],
  "otlp_endpoint": "http://localhost:4317",
  "otlp": {
    "protocol": "grpc",
    "timeout": "10s",
    "headers": {
      "Authorization": "Bearer token123"
    },
    "insecure": true
  },
  "development": false
}
```

## 🎯 OTLP 智能配置

### 配置优先级规则

Option 包实现了智能的 OTLP 配置冲突处理：

```go
// 1. 扁平化配置优先（简单场景）
opt := &option.LogOption{
    OTLPEndpoint: "http://localhost:4317", // 自动启用 OTLP
}

// 2. 明确禁用覆盖自动启用
opt := &option.LogOption{
    OTLPEndpoint: "http://localhost:4317",
    OTLP: &option.OTLPOption{
        Enabled: &[]bool{false}[0], // 明确禁用，优先级更高
    },
}

// 3. 嵌套配置（高级场景）
opt := &option.LogOption{
    OTLP: &option.OTLPOption{
        Enabled:  &[]bool{true}[0],
        Endpoint: "http://advanced:4317",
        Protocol: "http",
        Headers: map[string]string{
            "X-Custom": "value",
        },
    },
}
```

### 智能启用逻辑

```go
// 检查 OTLP 是否启用
if opt.IsOTLPEnabled() {
    fmt.Println("OTLP 已启用，端点:", opt.OTLP.Endpoint)
} else {
    fmt.Println("OTLP 未启用或配置不完整")
}

// 获取有效端点
endpoint := ""
if opt.OTLPEndpoint != "" {
    endpoint = opt.OTLPEndpoint // 扁平化优先
} else if opt.OTLP != nil {
    endpoint = opt.OTLP.Endpoint // 嵌套配置
}
```

## 📊 配置场景

### 生产环境配置

```go
func ProductionConfig() *option.LogOption {
    return &option.LogOption{
        Engine:            "zap",           // 高性能引擎
        Level:             "INFO",          // 生产级别
        Format:            "json",          // 结构化输出
        OutputPaths:       []string{"/var/log/app.log"},
        Development:       false,           // 生产模式
        DisableCaller:     false,           // 保留调用者信息
        DisableStacktrace: false,           // 保留错误堆栈
        
        // OTLP 生产配置
        OTLPEndpoint: "https://otlp.company.com:4317",
        OTLP: &option.OTLPOption{
            Protocol: "grpc",
            Timeout:  30 * time.Second,
            Headers: map[string]string{
                "Authorization": "Bearer " + os.Getenv("OTLP_TOKEN"),
            },
            Insecure: false, // 生产环境使用安全连接
        },
    }
}
```

### 开发环境配置

```go
func DevelopmentConfig() *option.LogOption {
    return &option.LogOption{
        Engine:      "slog",            // 标准库引擎
        Level:       "DEBUG",           // 调试级别
        Format:      "console",         // 易读格式
        OutputPaths: []string{"stdout"}, // 控制台输出
        Development: true,              // 开发模式
        
        // 本地 OTLP 测试
        OTLPEndpoint: "http://localhost:4317",
        OTLP: &option.OTLPOption{
            Protocol: "grpc",
            Insecure: true, // 本地测试允许不安全连接
        },
    }
}
```

### 测试环境配置

```go
func TestConfig() *option.LogOption {
    return &option.LogOption{
        Engine:      "slog",
        Level:       "ERROR",           // 只记录错误
        Format:      "json", 
        OutputPaths: []string{"stderr"}, // 错误输出
        Development: true,
        
        // 测试时禁用 OTLP
        OTLP: &option.OTLPOption{
            Enabled: &[]bool{false}[0], // 明确禁用
        },
    }
}
```

## 🔍 配置验证

### 基础验证

```go
opt := &option.LogOption{
    Engine: "unknown", // 无效引擎
    Level:  "INVALID", // 无效级别
}

err := opt.Validate()
if err != nil {
    fmt.Println("配置错误:", err)
    // 配置会自动修正为合理默认值
    fmt.Println("修正后引擎:", opt.Engine) // "slog"
}
```

### OTLP 配置检查

```go
// 检查配置状态
func checkOTLPConfig(opt *option.LogOption) {
    if opt.IsOTLPEnabled() {
        fmt.Printf("✅ OTLP 已启用: %s\n", opt.OTLP.Endpoint)
        fmt.Printf("   协议: %s\n", opt.OTLP.Protocol)
        fmt.Printf("   超时: %v\n", opt.OTLP.Timeout)
    } else {
        fmt.Println("❌ OTLP 未启用")
        
        if opt.OTLP != nil && opt.OTLP.Enabled != nil && !*opt.OTLP.Enabled {
            fmt.Println("   原因: 明确禁用")
        } else if opt.OTLP == nil || opt.OTLP.Endpoint == "" {
            fmt.Println("   原因: 缺少端点配置")
        }
    }
}
```

## 🧪 高级用法

### 动态配置合并

```go
// 基础配置
base := option.DefaultLogOption()

// 环境特定配置
override := &option.LogOption{
    Level: "DEBUG",
    OTLPEndpoint: "http://dev:4317",
}

// 合并配置（需要自实现合并逻辑）
mergedOpt := mergeConfigs(base, override)
```

### 条件配置

```go
func createConfig(env string) *option.LogOption {
    opt := option.DefaultLogOption()
    
    switch env {
    case "production":
        opt.Engine = "zap"
        opt.Level = "INFO"
        opt.Format = "json"
        opt.OTLPEndpoint = os.Getenv("PROD_OTLP_ENDPOINT")
        
    case "development":
        opt.Level = "DEBUG"
        opt.Format = "console"
        opt.Development = true
        opt.OTLPEndpoint = "http://localhost:4317"
        
    case "test":
        opt.Level = "ERROR"
        opt.OTLP = &option.OTLPOption{
            Enabled: &[]bool{false}[0],
        }
    }
    
    return opt
}
```

## 📋 测试支持

```bash
# 运行配置包测试
go test github.com/kart-io/logger/option -v

# 运行验证测试
go test github.com/kart-io/logger/option -run TestValidation

# 测试覆盖率
go test github.com/kart-io/logger/option -cover
```

## 🔗 相关资源

- [`core`](../core/) - 核心接口定义，Level 类型验证
- [`engines/zap`](../engines/zap/) - Zap 引擎配置应用
- [`engines/slog`](../engines/slog/) - Slog 引擎配置应用
- [`factory`](../factory/) - 基于配置创建日志器
- [`config`](../config/) - 高级配置管理和多源合并
- [Spf13/pflag](https://github.com/spf13/pflag) - 命令行标志库

## ⚠️ 注意事项

### 配置优先级

1. **扁平化优先**: `OTLPEndpoint` 优先于 `OTLP.Endpoint`
2. **明确禁用优先**: `OTLP.Enabled = false` 覆盖所有自动启用逻辑
3. **端点必需**: OTLP 启用需要有效的端点配置

### 类型处理

```go
// ✅ 正确：使用指针处理三态布尔
enabled := true
opt.OTLP.Enabled = &enabled

// ❌ 避免：直接赋值丢失 nil 状态
opt.OTLP.Enabled = true // 编译错误
```

### 配置验证

1. **引擎验证**: 无效引擎自动回退到 "slog"
2. **级别验证**: 使用 `core.ParseLevel` 严格验证
3. **OTLP 解析**: `Validate()` 自动应用智能配置逻辑
4. **默认值填充**: 缺失的配置项自动使用合理默认值

## 🚀 最佳实践

### 配置组织

```go
// ✅ 推荐：分环境配置函数
func NewConfig(env string) *option.LogOption {
    opt := option.DefaultLogOption()
    
    // 环境特定修改
    switch env {
    case "prod":
        return ProductionConfig()
    case "dev":
        return DevelopmentConfig() 
    default:
        return opt
    }
}
```

### 配置验证

```go
// ✅ 推荐：始终验证配置
opt := createConfig()
if err := opt.Validate(); err != nil {
    log.Fatalf("配置验证失败: %v", err)
}
```

### OTLP 端点检查

```go
// ✅ 推荐：检查 OTLP 状态
if opt.IsOTLPEnabled() {
    log.Printf("OTLP 追踪已启用: %s", opt.OTLP.Endpoint)
}
```

选择 Option 包，为你的应用提供灵活、智能的日志配置管理！ 🚀