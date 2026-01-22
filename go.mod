module github.com/hechh/framework

go 1.24.2

require (
	github.com/golang/protobuf v1.5.4
	github.com/gorilla/websocket v1.5.3
	github.com/hechh/library v0.0.1
	github.com/nats-io/nats.go v1.48.0
	github.com/spf13/cast v1.10.0
	go.etcd.io/etcd/client/v3 v3.6.7
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.5 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.etcd.io/etcd/api/v3 v3.6.7 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.7 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/grpc v1.78.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	google.golang.org/genproto => google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb
	google.golang.org/genproto/googleapis/api => google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb
	google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb
)
