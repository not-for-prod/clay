package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	contextPackage       = protogen.GoImportPath("context")
	grpcPackage          = protogen.GoImportPath("google.golang.org/grpc")
	embedPackage         = protogen.GoImportPath("embed")
	httptransportPackage = protogen.GoImportPath("github.com/not-for-prod/clay/transport/httptransport")
	transportPackage     = protogen.GoImportPath("github.com/not-for-prod/clay/transport")
	runtimePackage       = protogen.GoImportPath("github.com/grpc-ecosystem/grpc-gateway/v2/runtime")
)

func main() {
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(
		func(p *protogen.Plugin) error {
			p.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

			for _, f := range p.Files {
				if !f.Generate {
					continue
				}
				generate(p, f)
			}

			return nil
		},
	)
}

func generate(p *protogen.Plugin, f *protogen.File) {
	if len(f.Services) != 1 {
		return
	}

	service := f.Services[0]
	descName := service.GoName + "ServiceDesc"

	g := p.NewGeneratedFile(f.GeneratedFilenamePrefix+".pb.goclay.go", f.GoImportPath)
	g.P("// Code generated by protoc-gen-goclay. DO NOT EDIT.")
	g.P()
	g.P("package ", f.GoPackageName)
	g.P()
	g.Import(embedPackage)
	g.P()
	g.P("//go:embed ", trimPathAndExt(f.Proto.GetName()), ".swagger.json")
	g.P("var Swagger []byte")
	g.P()
	g.P("// ", descName, " is a descriptor/registrator for the ", service.GoName, "Server.")
	g.P("type ", descName, " struct {")
	g.P("svc ", service.GoName, "Server")
	g.P("opts ", httptransportPackage.Ident("DescOptions"))
	g.P("}")
	g.P()
	g.P("// New", descName, " creates new registrator for the "+service.GoName+"Server.")
	g.P("// It implements httptransport.ConfigurableServiceDesc as well.")
	g.P("func New", descName, "(i ", service.GoName, "Server) ", "*", descName, " {")
	g.P("return &", descName, "{svc: i}")
	g.P("}")
	g.P()
	g.P("// RegisterGRPC implements service registrator interface.")
	g.P("func(d *", descName, ") RegisterGRPC(s *", g.QualifiedGoIdent(grpcPackage.Ident("Server")), ") {")
	g.P("Register", service.GoName, "Server(s, d.svc)")
	g.P("}")
	g.P()
	g.P("// Apply applies passed options.")
	g.P("func(d *", descName, ") Apply(oo ...", transportPackage.Ident("DescOption"), ") {")
	g.P("for _, o := range oo {")
	g.P("o.Apply(&d.opts)")
	g.P("}")
	g.P("}")
	g.P()
	g.P("// SwaggerDef returns this file's Swagger definition.")
	g.P("func(d *", descName, ") SwaggerDef() []byte {")
	g.P(`return Swagger`)
	g.P("}")
	g.P()
	g.P("// RegisterHTTP registers this service's HTTP handlers/bindings.")
	g.P("func(w *", descName, ") RegisterHTTP(")
	g.P("ctx ", g.QualifiedGoIdent(contextPackage.Ident("Context")), ",")
	g.P("mux *", g.QualifiedGoIdent(runtimePackage.Ident("ServeMux")), ",")
	g.P(") error {")
	g.P("return Register", service.GoName, "HandlerServer(ctx, mux, w)")
	g.P("}")
	g.P()
	g.P("// Wrap all http methods with interceptor support")
	g.P()
	// Wrapper method implementations.
	for _, method := range service.Methods {
		if !method.Desc.IsStreamingClient() && !method.Desc.IsStreamingServer() {
			genServerMethod(g, method)
		}
	}
}

func genServerMethod(
	g *protogen.GeneratedFile,
	method *protogen.Method,
) {
	service := method.Parent
	descName := service.GoName + "ServiceDesc"

	g.P("func (w *", descName, ") ", method.GoName, "(ctx ", contextPackage.Ident("Context"), ", in *",
		method.Input.GoIdent, ") (*",
		method.Output.GoIdent, ", error) {")
	g.P("if w.opts.UnaryInterceptor == nil { return w.svc.", method.GoName, "(ctx, in) }")
	g.P("info := &", grpcPackage.Ident("UnaryServerInfo"), "{")
	g.P("Server: w,")
	g.P("FullMethod: ", strconv.Quote(fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())), ",")
	g.P("}")
	g.P("handler := func(ctx ", contextPackage.Ident("Context"), ", req interface{}) (interface{}, error) {")
	g.P("return w.svc.", method.GoName, "(ctx, req.(*", method.Input.GoIdent, "))")
	g.P("}")
	g.P("resp, err := w.opts.UnaryInterceptor(ctx, in, info, handler)")
	g.P("if err != nil || resp == nil {")
	g.P("return nil, err")
	g.P("}")
	g.P("return resp.(*", method.Output.GoIdent, "), err")
	g.P("}")
	g.P()
}

func trimPathAndExt(fName string) string {
	f := filepath.Base(fName)
	ext := filepath.Ext(f)
	return f[:len(f)-len(ext)]
}
