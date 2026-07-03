package consts

var (
	BuildDate = "unknown"
)

func init() {
	buildDateContent, _ := Files.ReadFile("build_date")
	if len(buildDateContent) > 0 {
		BuildDate = string(buildDateContent)
	}
}
