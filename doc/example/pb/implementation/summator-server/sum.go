package summator_server

import (
	context "context"

	example "github.com/utrack/clay/doc/example/pb"
)

func (i *Implementation) Sum(ctx context.Context, argb *example.SumRequest) (*example.SumResponse, error) {
	panic("implement me")
}
