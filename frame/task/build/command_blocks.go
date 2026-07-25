package build

import (
	"context"
	"fmt"
	"strings"
	"sync"

	packet_pb "github.com/HonLaderDev/HonLader-core-api/pb/minecraft/protocol/packet"
	mc_packet "github.com/HonLaderDev/mousetunnel/minecraft/protocol/packet"
)

const commandBlocksEnabledRuleName = "commandBlocksEnabled"

// startCommandBlocksDisabledGuard 保证构建期间命令方块不会运行。
func (b *BuildTask) startCommandBlocksDisabledGuard(ctx context.Context) (func(), error) {
	if b.DisableAutoCommandBlocksDisabled {
		return func() {}, nil
	}

	mu := new(sync.Mutex)
	if err := b.ensureCommandBlocksDisabled(ctx, mu); err != nil {
		return nil, fmt.Errorf("BuildTask.startCommandBlocksDisabledGuard: %w", err)
	}

	listenerID, err := b.frame.Client().Resources().PacketListener().ListenPacket(
		ctx,
		[]uint32{mc_packet.IDGameRulesChanged},
		func(pk *packet_pb.Packet, err error) {
			if err != nil {
				b.publish(EventNameRunCommandBlocksGuardFailed, err)
				return
			}
			if ctx.Err() != nil {
				return
			}
			b.handleCommandBlocksGameRulesChanged(ctx, mu, pk)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("BuildTask.startCommandBlocksDisabledGuard: listen game rules changed: %w", err)
	}

	return func() {
		_ = b.frame.Client().Resources().PacketListener().DestroyListener(context.Background(), listenerID)
	}, nil
}

func (b *BuildTask) ensureCommandBlocksDisabled(ctx context.Context, mu *sync.Mutex) error {
	mu.Lock()
	defer mu.Unlock()

	enabled, existed, err := b.commandBlocksEnabled(ctx)
	if err != nil {
		return fmt.Errorf("ensureCommandBlocksDisabled: %w", err)
	}
	if existed && !enabled {
		return nil
	}

	if err := b.sendSettingsCommand(ctx, "gamerule commandBlocksEnabled false", false); err != nil {
		return fmt.Errorf("ensureCommandBlocksDisabled: %w", err)
	}
	b.publish(EventNameRunCommandBlocksDisableFinish)
	return nil
}

func (b *BuildTask) commandBlocksEnabled(ctx context.Context) (bool, bool, error) {
	rule, existed, err := b.frame.Client().Resources().UQHolder().World().GetGameRule(ctx, commandBlocksEnabledRuleName)
	if err != nil || !existed {
		return true, existed, err
	}
	return !strings.EqualFold(strings.TrimSpace(rule.GetValue()), "false"), true, nil
}

func (b *BuildTask) handleCommandBlocksGameRulesChanged(ctx context.Context, mu *sync.Mutex, pk *packet_pb.Packet) {
	if pk == nil || pk.GetValue() == nil {
		return
	}

	changed := new(packet_pb.GameRulesChanged)
	if err := pk.GetValue().UnmarshalTo(changed); err != nil {
		b.publish(EventNameRunCommandBlocksGuardFailed, err)
		return
	}

	for _, rule := range changed.GetGameRules() {
		if rule == nil || !strings.EqualFold(rule.GetName(), commandBlocksEnabledRuleName) {
			continue
		}
		if rule.GetValue().GetBoolValue() {
			b.publish(EventNameRunCommandBlocksReenabled)
			if err := b.ensureCommandBlocksDisabled(ctx, mu); err != nil {
				b.publish(EventNameRunCommandBlocksGuardFailed, err)
			}
		}
		return
	}
}
