//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		time.Local = loc
	}

	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	hashOutput, err := hashCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取 commit 哈希失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile("commit_hash", []byte(strings.TrimSpace(string(hashOutput))), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "写入 commit_hash 文件失败: %v\n", err)
		os.Exit(1)
	}

	numCmd := exec.Command("git", "rev-list", "--count", "HEAD")
	numOutput, err := numCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取提交次数失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile("commit_num", []byte(strings.TrimSpace(string(numOutput))), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "写入 commit_num 文件失败: %v\n", err)
		os.Exit(1)
	}

	buildTime := time.Now().Format("2006-01-02 15:04:05 Monday")
	if err := os.WriteFile("build_date", []byte(buildTime), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "写入 build_date 文件失败: %v\n", err)
		os.Exit(1)
	}

	generateGitLogInfo()
}

type CommitInfo struct {
	Hash        string `json:"hash"`
	Author      string `json:"author"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

func generateGitLogInfo() {
	logCmd := exec.Command("git", "log", "--pretty=format:%H|%an|%ad|%s", "--date=format:%Y-%m-%d %H:%M:%S")
	logOutput, err := logCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取 git log 失败: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(logOutput)), "\n")
	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}
		commits = append(commits, CommitInfo{
			Hash:        parts[0],
			Author:      parts[1],
			Date:        parts[2],
			Description: parts[3],
		})
	}

	jsonData, err := json.Marshal(commits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "序列化 git log 失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile("commit_log.json", jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "写入 commit_log.json 文件失败: %v\n", err)
		os.Exit(1)
	}
}
