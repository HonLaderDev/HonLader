package data

import (
	"testing"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/spf13/afero"
)

func TestLatestServerConfig(t *testing.T) {
	manager := NewDataManager(memoryStorage{fs: afero.NewMemMapFs()})
	config := define.ServerConfig{
		Metadata:       define.Metadata{Name: "server"},
		ServerCode:     "123",
		ServerPassword: "password",
	}

	if err := manager.SaveLatestServerConfig(config); err != nil {
		t.Fatalf("save latest server config: %v", err)
	}
	loaded, exists, err := manager.LoadLatestServerConfig()
	if err != nil {
		t.Fatalf("load latest server config: %v", err)
	}
	if !exists {
		t.Fatal("latest server config does not exist")
	}
	if loaded.Metadata.Name != config.Metadata.Name || loaded.ServerCode != config.ServerCode || loaded.ServerPassword != config.ServerPassword {
		t.Fatalf("latest server config = %#v, want %#v", loaded, config)
	}
}

func TestDeleteServerConfigDeletesLatestServerConfig(t *testing.T) {
	manager := NewDataManager(memoryStorage{fs: afero.NewMemMapFs()})
	config := define.ServerConfig{
		Metadata:   define.Metadata{Name: "server"},
		ServerCode: "123",
	}

	if err := manager.SaveServerConfig(config); err != nil {
		t.Fatalf("save server config: %v", err)
	}
	if err := manager.SaveLatestServerConfig(config); err != nil {
		t.Fatalf("save latest server config: %v", err)
	}
	if deleted, err := manager.DeleteServerConfig(config.Metadata.Name); err != nil {
		t.Fatalf("delete server config: %v", err)
	} else if !deleted {
		t.Fatal("server config was not deleted")
	}

	if _, exists, err := manager.LoadLatestServerConfig(); err != nil {
		t.Fatalf("load latest server config: %v", err)
	} else if exists {
		t.Fatal("latest server config still exists")
	}
}
