package server

import (
	"context"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	structure_utils "github.com/HonLaderDev/HonLader/utils/structure"
	wsdefine "github.com/HonLaderDev/WaterStructure/define"
	waterstructure "github.com/HonLaderDev/WaterStructure/structure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// formatExtensions 保存文件选择器可使用的扩展名提示。真正的格式判断仍由
// StructureFromFile 根据扩展名、文件头和实际解析结果完成，不能信任此映射。
var formatExtensions = map[string][]string{
	waterstructure.NameSchematic:    {".schematic"},
	waterstructure.NameSchemV1:      {".schem"},
	waterstructure.NameSchemV2:      {".schem"},
	waterstructure.NameLitematic:    {".litematic"},
	waterstructure.NameMCStructure:  {".mcstructure"},
	waterstructure.NameMCWorld:      {".mcworld", ".zip"},
	waterstructure.NameBDX:          {".bdx"},
	waterstructure.NameConstruction: {".construction"},
	waterstructure.NameAxiomBP:      {".bp"},
	waterstructure.NameMCFunction:   {".mcfunction", ".txt"},
	waterstructure.NameKBDX:         {".kbdx"},
	waterstructure.NameIBImport:     {".ibi"},
	waterstructure.NameMianYangV3:   {".building"},
	waterstructure.NameMianYangV4:   {".buildingx"},
	waterstructure.NameGangBanV7:    {".reb"},
	waterstructure.NameFuHongV5:     {".fhbuild"},
	waterstructure.NameBDS:          {".bds"},
	waterstructure.NameSIBI:         {".sibi"},
	waterstructure.NameBCF:          {".bcf"},
	waterstructure.NameTIBI:         {".tibi"},
	waterstructure.NameCovStructure: {".covstructure"},
	waterstructure.NameNexusNP:      {".np"},
}

// ListSupportedFormats 返回运行时实际注册的格式，避免 Web 与服务端各自维护
// 一份最终必然漂移的格式列表。
func (s *Server) ListSupportedFormats(context.Context, *emptypb.Empty) (*apiv1.ListSupportedFormatsResponse, error) {
	formats := make([]*apiv1.StructureFormat, 0, len(waterstructure.StructureIDPool))
	for id, factory := range waterstructure.StructureIDPool {
		formats = append(formats, structureFormat(uint32(id), factory().Name()))
	}
	sort.Slice(formats, func(i, j int) bool { return formats[i].GetName() < formats[j].GetName() })
	return &apiv1.ListSupportedFormatsResponse{Formats: formats}, nil
}

// InspectStructure 解析文件并返回格式与空间尺寸，不创建中间世界。
func (s *Server) InspectStructure(ctx context.Context, req *apiv1.InspectStructureRequest) (*apiv1.StructureInfo, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "inspect structure request is required")
	}
	path, err := absPath(req.GetSourcePath(), "source_path")
	if err != nil {
		return nil, err
	}
	return inspectStructure(ctx, path)
}

// ConvertToMCWorld 将 WaterStructure 支持的任意输入转换为 mcworld，并以较低
// 频率推送进度，避免大建筑的逐区块回调压满 gRPC 流和上层 SSE。
func (s *Server) ConvertToMCWorld(req *apiv1.ConvertStructureRequest, stream apiv1.StructureService_ConvertToMCWorldServer) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "convert structure request is required")
	}
	source, err := absPath(req.GetSourcePath(), "source_path")
	if err != nil {
		return err
	}
	destination, err := absPath(req.GetDestinationPath(), "destination_path")
	if err != nil {
		return err
	}
	if err := stream.Send(&apiv1.StructureConversionEvent{
		Phase: apiv1.StructureConversionPhase_STRUCTURE_CONVERSION_PHASE_STARTED,
	}); err != nil {
		return err
	}

	converter := structure_utils.NewStructureConverter()
	progress := &conversionProgress{stream: stream, lastSent: time.Now()}
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertParsed, progress.onParsed)
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertProgressInit, progress.onInit)
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertProgress, progress.onProgress)

	result, err := converter.ConvertToMCWorldWithName(source, destination, strings.TrimSpace(req.GetWorldName()))
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "convert structure: %v", err)
	}
	if progress.sendErr != nil {
		return progress.sendErr
	}
	info, err := inspectStructure(stream.Context(), destination)
	if err != nil {
		return status.Errorf(codes.Internal, "inspect converted world: %v", err)
	}
	return stream.Send(&apiv1.StructureConversionEvent{
		Phase:      apiv1.StructureConversionPhase_STRUCTURE_CONVERSION_PHASE_COMPLETED,
		Current:    progress.total,
		Total:      progress.total,
		Info:       info,
		OutputPath: result.Path,
	})
}

// conversionProgress 串行化转换器回调，并限制进度消息频率。
type conversionProgress struct {
	mu       sync.Mutex
	stream   apiv1.StructureService_ConvertToMCWorldServer
	info     *apiv1.StructureInfo
	current  int64
	total    int64
	lastSent time.Time
	sendErr  error
}

func (p *conversionProgress) onParsed(name string, size wsdefine.Size) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info = structureInfo(name, size.Width, size.Height, size.Length, "")
	p.sendLocked(apiv1.StructureConversionPhase_STRUCTURE_CONVERSION_PHASE_PARSED)
}

func (p *conversionProgress) onInit(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = int64(total)
	p.sendLocked(apiv1.StructureConversionPhase_STRUCTURE_CONVERSION_PHASE_CONVERTING)
}

func (p *conversionProgress) onProgress() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current++
	if p.current == p.total || time.Since(p.lastSent) >= 100*time.Millisecond {
		p.sendLocked(apiv1.StructureConversionPhase_STRUCTURE_CONVERSION_PHASE_CONVERTING)
	}
}

func (p *conversionProgress) sendLocked(phase apiv1.StructureConversionPhase) {
	if p.sendErr != nil {
		return
	}
	p.lastSent = time.Now()
	p.sendErr = p.stream.Send(&apiv1.StructureConversionEvent{
		Phase: phase, Current: p.current, Total: p.total, Info: p.info,
	})
}

func inspectStructure(ctx context.Context, path string) (*apiv1.StructureInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "open structure: %v", err)
	}
	defer file.Close()
	item, err := waterstructure.StructureFromFile(file)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "inspect structure: %v", err)
	}
	defer item.Close()
	size := item.GetSize()
	return structureInfo(item.Name(), size.Width, size.Height, size.Length, structure_utils.DefaultWorldName(path)), nil
}

func structureInfo(name string, width, height, length int, worldName string) *apiv1.StructureInfo {
	return &apiv1.StructureInfo{
		Format:             structureFormatByName(name),
		Width:              int32(width),
		Height:             int32(height),
		Length:             int32(length),
		SuggestedWorldName: worldName,
		SuggestedStartPos:  &apiv1.BlockPos{X: 0, Y: -64, Z: 0},
		SuggestedEndPos:    &apiv1.BlockPos{X: int32(width - 1), Y: int32(height - 65), Z: int32(length - 1)},
	}
}

func structureFormatByName(name string) *apiv1.StructureFormat {
	for id, factory := range waterstructure.StructureIDPool {
		if factory().Name() == name {
			return structureFormat(uint32(id), name)
		}
	}
	return structureFormat(0, name)
}

func structureFormat(id uint32, name string) *apiv1.StructureFormat {
	extensions := append([]string(nil), formatExtensions[name]...)
	if len(extensions) == 0 {
		extensions = []string{".json"}
	}
	return &apiv1.StructureFormat{
		Id: id, Name: name, Extensions: extensions, Experimental: name == waterstructure.NameMianYangV4,
	}
}
