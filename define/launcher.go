package define

// Launcher 聚合任务运行框架、存储和数据管理能力。
type Launcher interface {
	TaskFrame
	Storage
	DataManager() DataManager
}
