# artisan (tkeel-tool)

The artisan 是面向 tKeel 开发者的开发工具，方便快速生成框架代码。

## Getting Started
### Required
- [go](https://golang.org/dl/)
- [protoc](https://github.com/protocolbuffers/protobuf)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go)

### Quick Start
```
# 安装
go get -u github.com/tkeel-io/tkeel-interface/tool/cmd/artisan

# 创建项目模板
artisan new github.com/tkeel-io/helloworld

cd helloworld

# 下载必须的插件
make init

# 生成proto模板
artisan proto add api/helloworld/v1/helloworld.proto

# 下载必须的插件
make api

# 生成service模板
artisan proto service api/helloworld/v1/helloworld.proto -t pkg/service

# 生成server模板(此输出需要手工加入 cmd/helloworld/main.go 中)
artisan proto server api/helloworld/v1/helloworld.proto

# 生成 api 的 makedown 文件
artisan markdown -f api/apidocs.swagger.json  -t third_party/markdown-templates/ -o ./docs/API/Greeter -m all

# 运行程序
go run cmd/helloworld/main.go
```



