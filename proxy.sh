# git config --global https.proxy http://127.0.0.1:1086
# git config --global https.proxy https://127.0.0.1:1086
# git config --global --unset http.proxy
# git config --global --unset https.proxy

# 启用 Go Modules 功能
# go env -w GO111MODULE=on

# 配置 GOPROXY 环境变量，以下三选一

# 1. 官方
# go env -w  GOPROXY=https://goproxy.io

# 2. 七牛 CDN
# go env -w  GOPROXY=https://goproxy.cn

# 3. 阿里云
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/

# confirm goproxy
go env | grep GOPROXY
