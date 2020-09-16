module server

go 1.14

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.496 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/henryleu/vads v0.2.1
)

replace (
   github.com/henryleu/vads => ../..
)
