package export

// Name 是导出任务名称。
const Name = "Export"

func (e ExportTask) Name() string {
	return Name
}
