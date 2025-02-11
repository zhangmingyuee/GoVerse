# 声明伪目标，防止与同名文件冲突
.PHONY: all build run gotool clean help

# 定义生成的二进制文件名称
BINARY="bluebell"

# 默认目标：先运行 gotool，再编译生成二进制文件
all: gotool build

# build 目标：编译 Go 代码，生成 Linux 平台上的 amd64 二进制文件
build:
	# 设置环境变量，关闭 CGO，以保证编译出的二进制文件可移植
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}

# run 目标：直接运行 Go 代码（通常用于开发调试）
run:
	@go run ./

# gotool 目标：使用 Go 的工具对代码进行格式化和静态检查
gotool:
	# 格式化代码
	go fmt ./
	# 检查代码潜在问题
	go vet ./

# clean 目标：删除生成的二进制文件
clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

# help 目标：显示可用的命令及说明
help:
	@echo "make         - 格式化 Go 代码，并编译生成二进制文件"
	@echo "make build   - 编译 Go 代码，生成二进制文件"
	@echo "make run     - 直接运行 Go 代码"
	@echo "make clean   - 移除二进制文件"
	@echo "make gotool  - 运行 Go 工具 'fmt' 和 'vet'"
