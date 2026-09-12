package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/HonLaderDev/HonLaderAuth-client/client"
	"github.com/HonLaderDev/HonLaderAuth-protocol/api"
	"github.com/HonLaderDev/mousetunnel/minecraft"
	mc_protocol "github.com/HonLaderDev/mousetunnel/minecraft/protocol"
	mc_packet "github.com/HonLaderDev/mousetunnel/minecraft/protocol/packet"
	"github.com/HonLaderDev/tanlobbyclient/tanlobby"
	tan_login "github.com/HonLaderDev/tanlobbyclient/tanlobby/protocol/login"
	tan_packet "github.com/HonLaderDev/tanlobbyclient/tanlobby/protocol/packet"
)

const tanLobbyProtocolIDPE = 42

func main() {
	ctx := context.Background()
	apiKey := envOr("HLA_API_KEY", "test")
	serverCode := envOr("HLA_SERVER_CODE", "9499721")
	serverPassword := envOr("HLA_SERVER_PASSWORD", "")

	hlaClient := client.HonLaderAuthClientConfig{
		BaseURL: "http://127.0.0.1:8080",
		APIKey:         apiKey,
		ServerCode:     serverCode,
		ServerPassword: serverPassword,
	}.New()
	auth := hlaClient.Authenticator()

	fmt.Println("requesting tan lobby create access")
	tanResp, err := auth.CreateTanLobby(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("creating tan lobby room")
	roomConn, roomID, err := createRoom(ctx, tanResp)
	if err != nil {
		log.Fatal(err)
	}
	defer roomConn.Close()

	listener, err := minecraft.ListenConfig{
		AuthenticationDisabled: true,
		ErrorLog:               slog.Default(),
		Compression:            mc_packet.NewNeteaseCompression(),
		AcceptedProtocols: []minecraft.Protocol{
			minecraft.NewBedrockProtocol(mc_protocol.ProfileNetease1v21v124),
		},
		PongProtocol: minecraft.NewBedrockProtocol(mc_protocol.ProfileNetease1v21v124),
		PacketFunc: func(header mc_packet.Header, payload []byte, src, dst net.Addr) {
			fmt.Printf("mc packet id=%d src=%s dst=%s bytes=%d\n", header.PacketID, src, dst, len(payload))
		},
	}.Listen("raknet8", "0.0.0.0:19132")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("room id:", roomID)
	fmt.Println("room listener:", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go dumpConn(conn)
	}
}

func createRoom(ctx context.Context, response *api.TanLobbyCreateResponse) (minecraft.NetworkListener, uint32, error) {
	if response == nil || response.GetInfo() == nil || response.GetBot() == nil {
		return nil, 0, fmt.Errorf("create room: invalid tan lobby response")
	}
	roomInfo := response.GetInfo()
	bot := response.GetBot()
	loginData, err := tan_login.NewLoginData(
		uint32(bot.GetUid()),
		roomInfo.GetRaknetRand(),
		roomInfo.GetRaknetAesRand(),
		bot.GetName(),
		tan_login.PlatformPE,
		uint32(bot.GetUid()),
		roomInfo.GetEncryptKeyBytes(),
		roomInfo.GetDecryptKeyBytes(),
		roomInfo.GetSignalingServerInfo().GetAddress(),
		roomInfo.GetSignalingSeed(),
		roomInfo.GetSignalingTicket(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("create room: build login data: %w", err)
	}

	raknetAddr := roomInfo.GetRaknetServerInfo().GetAddress()
	if raknetAddr == "" {
		return nil, 0, fmt.Errorf("create room: empty raknet address")
	}
	tanConn, err := tanlobby.Dialer{
		ErrorLog:  slog.Default(),
		LoginData: loginData,
	}.DialContext(ctx, "raknet8", raknetAddr)
	if err != nil {
		return nil, 0, fmt.Errorf("create room: dial tan lobby: %w", err)
	}

	fmt.Println("connected tan lobby, sending create room request")
	roomListener, roomID, err := tanConn.CreateRoom(ctx, tanlobby.RoomConfig{
		Capacity:             8,
		Privacy:              tan_packet.PrivacyEveryoneCanSee,
		PlatformPrivacy:      tan_packet.PrivacyPlatformPCAndPE,
		Name:                 "",
		LevelID:              envOr("ROOM_LEVEL_ID", "world"),
		GameType:             0,
		Voice:                0,
		ProtocolID:           tanLobbyProtocolIDPE,
		HostMinecraftVersion: envOr("ROOM_HOST_VERSION", mc_protocol.Version1v21v124),
		Password:             envOr("ROOM_PASSWORD", ""),
		Slogan:               "",
		PvP:                  true,
		PlayerAuth:           tan_packet.PlayerAuthMember,
		EnableWebRTC:         true,
		OwnerPing:            tan_packet.OwnerPingGreen,
		PerfLv:               tan_packet.PerfLvFluent,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("create room: %w", err)
	}
	return roomListener, roomID, nil
}

func dumpConn(conn net.Conn) {
	defer conn.Close()
	mcConn, ok := conn.(*minecraft.Conn)
	if !ok {
		fmt.Printf("accepted %T\n", conn)
		return
	}
	fmt.Printf("identity: %#v\n", mcConn.IdentityData())
	fmt.Printf("client: %#v\n", mcConn.ClientData())
	for {
		pk, err := mcConn.ReadPacket()
		if err != nil {
			fmt.Println("read packet:", err)
			return
		}
		fmt.Printf("packet: %T %#v\n", pk, pk)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
