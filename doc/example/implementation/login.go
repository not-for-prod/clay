package sum

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	desc "github.com/utrack/clay/doc/example/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (i *SummatorImplementation) Login(ctx context.Context, req *desc.LoginRequest) (*desc.LoginResponse, error) {
	// Set cookie using gRPC metadata (Clay should handle this)
	md := metadata.New(
		map[string]string{
			"set-cookie": fmt.Sprintf(
				"%s=%s; Path=/; Max-Age=%d; HttpOnly; SameSite=Lax",
				"summator-session", uuid.NewString(), int(24*time.Hour.Seconds()),
			),
		},
	)

	// Send metadata as trailer (Clay should convert this to HTTP headers)
	err := grpc.SetHeader(ctx, md)
	if err != nil {
		return nil, err
	}

	return nil, err
}
