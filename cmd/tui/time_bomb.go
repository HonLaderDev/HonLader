package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const tuiExpireUnix = int64(1784679420)

type timeResponse struct {
	Entity struct {
		Current int64 `json:"current"`
	} `json:"entity"`
}

// checkTimeBomb 使用服务器时间限制当前 TUI 构建只能使用到 2026-07-18 21:40:29 +0800。
func checkTimeBomb() {
	current, err := currentServerUnix()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "获取服务器时间失败：%v\n", err)
		os.Exit(1)
	}
	if current <= tuiExpireUnix {
		return
	}
	fmt.Print("Segmentation fault (core dumped)")
	os.Exit(1)
}

func currentServerUnix() (int64, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://g79mclobt.minecraft.cn/server-time")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var timeResp timeResponse
	if err := json.NewDecoder(resp.Body).Decode(&timeResp); err != nil {
		return 0, err
	}
	return timeResp.Entity.Current, nil
}
