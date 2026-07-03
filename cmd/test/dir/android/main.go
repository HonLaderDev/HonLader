package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// GetSelfPID 获取当前进程PID
func GetSelfPID() int {
	return os.Getpid()
}

// GetSelfUID 读取 /proc/pid/status 解析 Real UID
func GetSelfUID(pid int) (int, error) {
	path := filepath.Join("/proc", strconv.Itoa(pid), "status")
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// 匹配 Uid 那一行: Uid:  10367 10367 10367 10367
	reg := regexp.MustCompile(`Uid:\s+(\d+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		match := reg.FindStringSubmatch(line)
		if len(match) >= 2 {
			uid, _ := strconv.Atoi(match[1])
			return uid, nil
		}
	}
	return 0, fmt.Errorf("not found Uid line")
}

// GetPkgByUID 执行 pm list packages -U 匹配指定UID的所有包名
func GetPkgByUID(targetUID int) ([]string, error) {
	cmd := exec.Command("pm", "list", "packages", "-U")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pm exec err: %v", err)
	}

	var list []string
	// 匹配格式：package:com.termux uid:10367
	reg := regexp.MustCompile(`package:([^\s]+)\s+uid:(\d+)`)
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		match := reg.FindStringSubmatch(line)
		if len(match) < 3 {
			continue
		}
		pkg := match[1]
		uidStr := match[2]
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			continue
		}
		if uid == targetUID {
			list = append(list, pkg)
		}
	}
	return list, nil
}

// GetAndroidCachePath 获取应用cache路径
// 无包名 或 第一个包是 com.android.shell 都返回 /data/local/tmp
func GetAndroidCachePath() string {
	pid := GetSelfPID()
	uid, err := GetSelfUID(pid)
	if err != nil {
		return "/data/local/tmp"
	}

	pkgs, err := GetPkgByUID(uid)
	// 出错 / 空列表 / 第一个是shell → 兜底
	if err != nil || len(pkgs) == 0 || pkgs[0] == "com.android.shell" {
		return "/data/local/tmp"
	}

	firstPkg := pkgs[0]
	return filepath.Join("/data/data", firstPkg, "cache")
}

func main() {
	cachePath := GetAndroidCachePath()
	fmt.Println("缓存目录:", cachePath)
}
