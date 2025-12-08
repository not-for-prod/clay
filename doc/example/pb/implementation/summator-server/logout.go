package summator_server

import (
	context "context"

	example "github.com/utrack/clay/doc/example/pb"
)

func (i *Implementation) Logout(ctx context.Context, argb *example.LogoutRequest) (*example.LogoutResponse, error) {
	panic("implement me")
}
