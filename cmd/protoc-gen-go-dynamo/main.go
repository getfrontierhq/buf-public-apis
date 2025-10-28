package main

import (
	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/getfrontierhq/buf-public-apis/internal/godynamo"
)

func main() {
	features := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	pgs.Init(
		pgs.SupportedFeatures(&features),
	).
		RegisterModule(godynamo.New()).
		RegisterPostProcessor(pgsgo.GoFmt()).
		Render()
}
