package transport

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type CompoundServiceDesc struct {
	svc []ServiceDesc
}

func NewCompoundServiceDesc(desc ...ServiceDesc) *CompoundServiceDesc {
	return &CompoundServiceDesc{svc: desc}
}

func (d *CompoundServiceDesc) RegisterGRPC(g *grpc.Server) {
	for _, svc := range d.svc {
		svc.RegisterGRPC(g)
	}
}

func (d *CompoundServiceDesc) RegisterHTTP(ctx context.Context, mux *runtime.ServeMux) error {
	for _, svc := range d.svc {
		if err := svc.RegisterHTTP(ctx, mux); err != nil {
			return err
		}
	}
	return nil
}

func (d *CompoundServiceDesc) SwaggerDef() []byte {
	j := &swagJoiner{}
	for _, svc := range d.svc {
		j.AddDefinition(svc.SwaggerDef())
	}
	return j.SumDefinitions()
}

func (d *CompoundServiceDesc) Apply(oo ...DescOption) {
	for _, ss := range d.svc {
		if s, ok := ss.(ConfigurableServiceDesc); ok {
			s.Apply(oo...)
		}
	}
}

type CompoundService struct {
	desc *CompoundServiceDesc
}

func NewCompoundService(svcs ...Service) *CompoundService {
	descs := make([]ServiceDesc, 0, len(svcs))

	for _, svc := range svcs {
		descs = append(descs, svc.GetDescription())
	}

	return &CompoundService{
		desc: NewCompoundServiceDesc(descs...),
	}
}

func (c *CompoundService) GetDescription() ServiceDesc {
	return c.desc
}
