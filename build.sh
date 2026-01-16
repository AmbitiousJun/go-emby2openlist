#!/bin/bash

# 删除dist
rm -rf ./dist

# 创建dist目录
mkdir -p dist

# 定义平台数组
platforms=("darwin/amd64" "darwin/arm64" "linux/386" "linux/arm" "linux/amd64" "linux/arm64" "windows/386" "windows/arm" "windows/amd64" "windows/arm64" "freebsd/386" "freebsd/arm" "freebsd/amd64" "freebsd/arm64")

# 循环编译并重命名可执行文件
for platform in "${platforms[@]}"
do
    # 分割平台字符串
    IFS='/' read -r -a platform_info <<< "$platform"

    # 构建平台
    GOOS=${platform_info[0]} GOARCH=${platform_info[1]}

    # 构建可执行文件名
    output_name="ge2o-${GOOS}_${GOARCH}"

    # Windows平台特殊处理，添加.exe后缀
    if [ "$GOOS" == "windows" ]; then
        output_name="$output_name.exe"
    fi

    # 编译
    CGO_ENABLED=0 GOOS=${platform_info[0]} GOARCH=${platform_info[1]} go build -tags=goexperiment.jsonv2 -ldflags="-X main.ginMode=release" -o "dist/$output_name" .

    # 赋予可执行权限
    chmod +x "dist/$output_name"

    echo "Built $output_name"
done

echo "Build process completed!"

