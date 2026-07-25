package build

import (
	"testing"
	"time"

	"go.uber.org/ratelimit"
)

// TestTakeCommandLimitAppliesLimiter 确认命令发送前的限速器会被实际消费。
func TestTakeCommandLimitAppliesLimiter(t *testing.T) {
	task := &BuildTask{limiter: ratelimit.New(20)}

	start := time.Now()
	task.takeCommandLimit()
	task.takeCommandLimit()
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Fatalf("takeCommandLimit did not apply limiter, elapsed=%s", elapsed)
	}
}
