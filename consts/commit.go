package consts

import "strconv"

var (
	CommitNum  = 0
	CommitHash = "unknown"
)

func init() {
	commitHashContent, _ := Files.ReadFile("commit_hash")
	if len(commitHashContent) > 0 {
		CommitHash = string(commitHashContent)
	}

	commitNumContent, _ := Files.ReadFile("commit_num")
	if len(commitNumContent) > 0 {
		num, err := strconv.Atoi(string(commitNumContent))
		if err == nil {
			CommitNum = num
		}
	}
}
