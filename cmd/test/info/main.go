package main

import (
	"fmt"

	"github.com/HonLaderDev/mousetunnel/minecraft/protocol"
	"github.com/HonLaderDev/mousetunnel/minecraft/protocol/login"
)

var identity = login.IdentityData{
	XUID:        "",
	Identity:    "65162105-6564-3e2f-a04b-a04c62cff741",
	DisplayName: "是只耶吧",
	NeteaseSid:  "984392:TanLobbyClient",
}

var client = login.ClientData{
	CapeID:                           "4672395235685216085",
	ClientRandomID:                   8225575651565618115,
	CurrentInputMode:                 2,
	DefaultInputMode:                 2,
	DeviceModel:                      "ONEPLUS PJF110",
	DeviceOS:                         protocol.DeviceAndroid,
	DeviceID:                         login.DeviceID("61f6a165440e4d75856826f2a353f4c0"),
	GameVersion:                      "1.21.120",
	LanguageCode:                     "zh_CN",
	PremiumSkin:                      true,
	SelfSignedID:                     "65162105-6564-3e2f-a04b-a04c62cff741",
	ServerAddress:                    ":0",
	SkinImageHeight:                  128,
	SkinImageWidth:                   128,
	SkinColour:                       "#0",
	SkinData:                         skinData,
	SkinGeometry:                     skinGeometry,
	SkinResourcePatch:                skinResourcePatch,
	SkinGeometryVersion:              "MC.0.0",
	SkinID:                           "c18e65aa-7b21-4637-9b63-8ad63622ef01.Custom61f6a165440e4d75856826f2a353f4c0",
	TrustedSkin:                      false,
	OverrideSkin:                     false,
	CompatibleWithClientSideChunkGen: true,
	MaxViewDistance:                  22,
	MemoryTier:                       4,
	PlatformType:                     1,
	GraphicsMode:                     1,
	GrowthLevel:                      47,
	ThirdPartyName:                   "是只耶吧",
	ArmSize:                          "wide",
}

func main() {
	fmt.Printf("identity: %#v\n", identity)
	fmt.Printf("client: %#v\n", client)
}
