package consts

import "encoding/json"

// CommitInfo 描述一次 Git 提交。
type CommitInfo struct {
	Hash        string `json:"hash"`
	Author      string `json:"author"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

var (
	CommitLog []CommitInfo
)

func init() {
	commitLogContent, _ := Files.ReadFile("commit_log.json")
	if len(commitLogContent) == 0 {
		CommitLog = []CommitInfo{}
		return
	}
	if err := json.Unmarshal(commitLogContent, &CommitLog); err != nil {
		CommitLog = []CommitInfo{}
	}
}

// GetCommitsSince 获取指定提交次数之后的提交。
func GetCommitsSince(lastCommitNum int) []CommitInfo {
	if lastCommitNum >= len(CommitLog) || lastCommitNum < 0 {
		return []CommitInfo{}
	}

	newCommitsCount := len(CommitLog) - lastCommitNum
	if newCommitsCount <= 0 {
		return []CommitInfo{}
	}
	return CommitLog[:newCommitsCount]
}
