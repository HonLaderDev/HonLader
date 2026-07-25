package main

import (
	"context"
	"log"
	"sort"

	packet_pb "github.com/HonLaderDev/HonLader-core-api/pb/minecraft/protocol/packet"
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/frame"
	"github.com/HonLaderDev/mousetunnel/minecraft/protocol/packet"
)

const cookie = `{"emulator":true,"is_guest":false,"mac_addr":"92f65770e558b6a1a3b1d5b64e320458","ram":1035337728,"rom":134208294912,"sauth_json":{"aim_info":"{\"aim\":\"127.0.0.1\",\"celluar_ip\":\"\",\"country\":\"CN\",\"is_vpn_enabled\":false,\"operator\":\"\",\"tz\":\"+0800\",\"tzid\":\"Asia/Shanghai\"}","app_channel":"4399com","client_login_sn":"42fd03ceb6f44f8f8b7ddabca5dad5fc","deviceid":"42fd03ceb6f44f8f8b7ddabca5dad5fc","gameid":"x19","get_access_token":"1","ip":"127.0.0.1","is_unisdk_guest":0,"login_channel":"4399com","platform":"ad","realname":"{\"realname_type\":0}","sdk_version":"1.0.0","sdkuid":"1391842071","sessionid":"1391842071|b695af2b6f2f9672838a055e7a585c88|44770||1d2a0c41a927e2400dad03f0bf9d7a4a|06a9ee393342b79c64a7337cd5c6c944|1785674131|4399","source_app_channel":"4399com","source_platform":"ad","tdid":"","udid":"7a837f0c9ddcc028"}}`

func main() {
	ctx := context.Background()
	tf := frame.TaskFrameConfig{Embedded: true}.New(nil)

	if err := tf.Connect(ctx, define.ConnectConfig{
		AuthServer: "http://127.0.0.1:8080/api/honlader",
		AuthToken:  cookie,
		ServerCode: "Player:2858926732",
		//ServerCode: "Player:2719712857",
		ServerPassword: "123456",
	}); err != nil {
		log.Fatal(err)
	}
	log.Println("entered")
	printRules(ctx, tf, "current")
	if _, err := tf.Client().Resources().PacketListener().ListenPacket(ctx, []uint32{packet.IDGameRulesChanged}, func(_ *packet_pb.Packet, err error) {
		if err != nil {
			log.Printf("listen game rules: %v", err)
			return
		}
		printRules(ctx, tf, "changed")
	}); err != nil {
		log.Fatal(err)
	}
	select {}
}

func printRules(ctx context.Context, tf *frame.TaskFrame, title string) {
	world := tf.Client().Resources().UQHolder().World()
	names, err := world.GetGameRuleNames(ctx)
	if err != nil {
		log.Printf("get game rule names: %v", err)
		return
	}
	sort.Strings(names)
	log.Printf("game rules %s total=%d", title, len(names))
	for _, name := range names {
		rule, ok, err := world.GetGameRule(ctx, name)
		if err != nil {
			log.Printf("get game rule %s: %v", name, err)
			continue
		}
		if !ok {
			continue
		}
		log.Printf("  %s=%s modifiable=%v", name, rule.GetValue(), rule.GetCanBeModifiedByPlayer())
	}
}
