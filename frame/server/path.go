package server

import (
	"path/filepath"

	"github.com/HonLaderDev/HonLader/define"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) currentStorage() define.Storage {
	if s.storage != nil {
		return s.storage
	}
	return nil
}

func absPath(path string, name string) (string, error) {
	if path == "" {
		return "", status.Errorf(codes.InvalidArgument, "%s is required", name)
	}
	result, err := filepath.Abs(path)
	if err != nil {
		return "", status.Errorf(codes.InvalidArgument, "invalid %s: %v", name, err)
	}
	return filepath.Clean(result), nil
}
