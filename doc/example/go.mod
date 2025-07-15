module github.com/utrack/clay/doc/example

go 1.24.1

replace github.com/not-for-prod/clay => ../../

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250625184727-c923a0c2a132.1
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.1
	github.com/not-for-prod/clay v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/peterbourgon/mergemap v0.0.1 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/swaggo/files v1.0.1 // indirect
	github.com/swaggo/http-swagger v1.3.4 // indirect
	github.com/swaggo/swag v1.16.4 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
