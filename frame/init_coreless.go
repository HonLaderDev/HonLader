//go:build coreless

package frame

import "fmt"

func (f *TaskFrame) initClient() error {
	if f.client == nil {
		return fmt.Errorf("TaskFrame.initClient: nil client")
	}
	return nil
}
