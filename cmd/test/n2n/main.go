package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

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
const authBaseURL = "http://127.0.0.1:8080"
const authAPIKey = "test"
const targetServerCode = "56242759"
const targetServerPassword = ""

func main() {
	ctx := context.Background()
	auth := newAuthClient().Authenticator()

	fmt.Println("requesting target server access")
	access, err := auth.Access(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connecting target server:", access.Address)
	serverConn, err := connectTargetServer(ctx, access)
	if err != nil {
		log.Fatal(err)
	}
	packetLog, err := newPacketJSONL("packet.jsonl")
	if err != nil {
		_ = serverConn.Close()
		log.Fatal(err)
	}
	defer packetLog.Close()

	fmt.Println("requesting tan lobby create access")
	tanResp, err := auth.CreateTanLobby(ctx)
	if err != nil {
		_ = serverConn.Close()
		log.Fatal(err)
	}

	fmt.Println("creating tan lobby room")
	roomConn, roomID, err := createRoom(ctx, tanResp)
	if err != nil {
		_ = serverConn.Close()
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
	}.Listen("raknet8", "0.0.0.0:19132")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("room id:", roomID)
	fmt.Println("waiting for exactly one player")
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	_ = listener.Close()

	mcConn, ok := conn.(*minecraft.Conn)
	if !ok {
		_ = conn.Close()
		_ = serverConn.Close()
		log.Fatalf("accepted unexpected connection type %T", conn)
	}
	if err := handleMITM(listener, mcConn, serverConn, packetLog); err != nil {
		log.Fatal(err)
	}
	select {}
}

func newAuthClient() *client.HonLaderAuthClient {
	return client.HonLaderAuthClientConfig{
		BaseURL:        authBaseURL,
		APIKey:         authAPIKey,
		ServerCode:     targetServerCode,
		ServerPassword: targetServerPassword,
	}.New()
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

func connectTargetServer(ctx context.Context, access minecraft.NeteaseAccess) (*minecraft.Conn, error) {
	return minecraft.Dialer{
		NeteaseAccess:              access,
		ErrorLog:                   slog.Default(),
		KeepXBLIdentityData:        true,
		DisconnectOnInvalidPackets: false,
		DisconnectOnUnknownPackets: false,
	}.DialNeteaseContext(ctx, "raknet8", access.Address)
}

func handleMITM(listener *minecraft.Listener, playerConn, serverConn *minecraft.Conn, packetLog *packetJSONL) error {
	fmt.Printf("accepted player: %s (%s)\n", playerConn.IdentityData().DisplayName, playerConn.IdentityData().Identity)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := playerConn.StartGame(serverConn.GameData()); err != nil {
			log.Printf("start player game: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := serverConn.DoSpawn(); err != nil {
			log.Printf("spawn on target server: %v", err)
		}
	}()
	wg.Wait()

	fmt.Println("mitm forwarding started")
	go forwardPackets("player -> server", "Client", playerConn, serverConn, packetLog, func(reason string) {
		_ = listener.Disconnect(playerConn, reason)
		_ = serverConn.Close()
	})
	go forwardPackets("server -> player", "Server", serverConn, playerConn, packetLog, func(reason string) {
		_ = serverConn.Close()
		_ = listener.Disconnect(playerConn, reason)
	})
	return nil
}

func forwardPackets(name, packetType string, src, dst *minecraft.Conn, packetLog *packetJSONL, closeBoth func(reason string)) {
	for {
		pk, err := src.ReadPacket()
		if err != nil {
			log.Printf("%s read stopped: %v", name, err)
			closeBoth("connection lost")
			return
		}
		if err := packetLog.Write(packetType, pk); err != nil {
			log.Printf("write packet log: %v", err)
		}
		if shouldDropPacket(packetType, pk) {
			continue
		}
		if err := dst.WritePacket(pk); err != nil {
			var disc minecraft.DisconnectError
			if errors.As(err, &disc) {
				closeBoth(disc.Error())
			} else {
				closeBoth("connection lost")
			}
			log.Printf("%s write stopped: %v", name, err)
			return
		}
	}
}

func shouldDropPacket(packetType string, pk mc_packet.Packet) bool {
	switch pk.ID() {
	case mc_packet.IDClientBoundMapItemData, mc_packet.IDServerBoundLoadingScreen, mc_packet.IDSyncActorProperty:
		return true
	case mc_packet.IDLevelEventGeneric,
		mc_packet.IDInventoryContent:
		return packetType == "Server"
	default:
		if pyRpc, ok := pk.(*mc_packet.PyRpc); ok {
			return shouldDropPyRpc(pyRpc)
		}
		return false
	}
}

func shouldDropPyRpc(pk *mc_packet.PyRpc) bool {
	values, ok := pk.Value.([]any)
	if !ok || len(values) == 0 {
		return false
	}
	name, ok := values[0].(string)
	if !ok {
		return false
	}
	switch name {
	case "GetMCPCheckNum", "SetMCPCheckNum", "GetStartType", "SetStartType":
		return true
	default:
		return false
	}
}

type packetJSONL struct {
	mu      sync.Mutex
	encoder *json.Encoder
	file    *os.File
}

func newPacketJSONL(path string) (*packetJSONL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &packetJSONL{
		encoder: json.NewEncoder(file),
		file:    file,
	}, nil
}

type packetLogEntry struct {
	Name        string `json:"Name"`
	ID          uint32 `json:"ID"`
	Type        string `json:"Type"`
	Data        any    `json:"Data"`
	EncodeError string `json:"EncodeError,omitempty"`
}

func (w *packetJSONL) Write(packetType string, pk mc_packet.Packet) error {
	entry := packetLogEntry{
		Name: fmt.Sprintf("%T", pk),
		ID:   pk.ID(),
		Type: packetType,
		Data: pk,
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.encoder.Encode(entry); err != nil {
		entry.Data = fmt.Sprintf("%#v", pk)
		entry.EncodeError = err.Error()
		return w.encoder.Encode(entry)
	}
	return nil
}

func (w *packetJSONL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
