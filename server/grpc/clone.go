package server

import (
	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	"google.golang.org/protobuf/proto"
)

func cloneConfig(v *apiv1.Config) *apiv1.Config {
	if v == nil {
		return nil
	}
	return proto.Clone(v).(*apiv1.Config)
}
