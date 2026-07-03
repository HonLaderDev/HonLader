//go:build !coreless

package frame

import (
	"fmt"

	"github.com/RedLaderDev/RedLader-core/frame/RedLaderCore/client"
)

func (f *TaskFrame) initClient() error {
	if f.client != nil {
		return nil
	}
	if !f.config.Embedded {
		return fmt.Errorf("TaskFrame.initClient: nil client")
	}

	f.client = client.New(nil)
	f.closer = f.client.Close
	return nil
}
