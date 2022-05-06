package opentelem

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type CoverageInfo struct {
	Count   *[]uint32
	Pos     *[]uint32
	NumStmt *[]uint16
}

// This is supposed to be a map <file_name> => <coverage_info>
// Since the cover tool only handles a single file.
// We need to find out how it can annotate multiple files when running tests.
var CoverageMap map[string]*CoverageInfo = make(map[string]*CoverageInfo)
var myProcessorInstance *MySpanProcessor

// This was an attempt to inject code from the coverage tool.
// The idea was that we would inject calls to this function to save the coverage structs.
// But you can't inject function calls outside a function.
func RecordCoverageMap(file string, coverage struct {
	Count   []uint32
	Pos     []uint32
	NumStmt []uint16
}) {
	if _, ok := CoverageMap[file]; ok {
		fmt.Printf("[DEBUG] Already have coverage info for %s\n", file)
		return
	}
	fmt.Printf("[DEBUG] New Coverage info for file %s\n", file)
	info := &CoverageInfo{Count: &coverage.Count, Pos: &coverage.Pos, NumStmt: &coverage.NumStmt}
	CoverageMap[file] = info
}

// I'm using a singleton because the SpanExporter needs access to the SpanProcessor
// because the span snapshots are recorded there (in this implementation)
func getProcessorSingleton() *MySpanProcessor {
	if myProcessorInstance == nil {
		myProcessorInstance = &MySpanProcessor{is_active: true, spanInfo: make(map[string]SpanCoverage)}
	}
	return myProcessorInstance
}

func GetCoverInfoForSpan() (*CoverageInfo, string, bool) {
	_, file, _, ok := runtime.Caller(3) // 3 comes from [GetCoverInfoForSpan (0), OnStart (1), some code in tracer (2), actual creation of Span (3)]
	fmt.Printf("[DEBUG] Caller file is %s\n", file)
	if !ok {
		fmt.Printf("[DEBUG][GetCoverInfoForSpan] runtime.Caller not ok\n")
		return nil, file, false
	}
	info, ok := CoverageMap[file]
	// ! We need to figure out how to get coverage info for multiple files.
	// And we don't access to the global struct here as well, cause it's in a different package
	return info, file, ok
}

// The idea (that doesn't quite work) was to create a snapshot of Coverage.Count when span is created
// And compare that to Coverage.Count when span is destroyed.
// But that fails because multithreading.
type SpanCoverage struct {
	refFile  string
	snapshot []uint32
}

// is_active was supposed to keep track if the Span is active or not. Not implemented.
// spanInfo is supposed be a map <spanID> => <coverageInfo>
type MySpanProcessor struct {
	is_active bool
	spanInfo  map[string]SpanCoverage
}

func (proc *MySpanProcessor) OnStart(parent context.Context, s trace.ReadWriteSpan) {
	for _, attr := range s.Attributes() {
		if attr.Key == "coverage" && attr.Value == attribute.BoolValue(true) {
			coverInfo, file, ok := GetCoverInfoForSpan()
			if !ok {
				fmt.Printf("[DEBUG][OnStart] Didn't find coverage info for Span file in records\n")
				return
			}
			var newArray []uint32 = make([]uint32, len(*coverInfo.Count))

			for i, v := range *coverInfo.Count {
				newArray[i] = v
			}
			spanInfo := SpanCoverage{refFile: file, snapshot: newArray}
			proc.spanInfo[s.SpanContext().SpanID().String()] = spanInfo
		}
	}
}

func (proc *MySpanProcessor) OnEnd(s trace.ReadOnlySpan) {
	if info, ok := proc.spanInfo[s.SpanContext().SpanID().String()]; ok {
		coverInfo, ok := CoverageMap[info.refFile]
		initialCoverage := info.snapshot
		if !ok {
			fmt.Printf("[DEBUG][OnEnd] Didn't find coverage info for Span file in records\n")
			return
		}
		for i, v := range *coverInfo.Count {
			if initialCoverage[i] == v {
				initialCoverage[i] = 0
			}
		}
	}
}

func (proc *MySpanProcessor) Shutdown(ctx context.Context) error {
	proc.is_active = false
	return nil
}

func (proc *MySpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

type CodecovExporter struct {
	wrappedExporter trace.SpanExporter
}

func (exp *CodecovExporter) ExportSpans(ctx context.Context, ss []trace.ReadOnlySpan) error {
	f, err := os.Create("coverage.txt")
	processor := getProcessorSingleton()
	if err == nil {
		for _, s := range ss {
			if spanInfo, ok := processor.spanInfo[s.SpanContext().SpanID().String()]; ok {
				fmt.Fprintf(f, "Coverage for Span %s.\nRef file: %s\nSnapshot length: %d\n", s.SpanContext().SpanID().String(), spanInfo.refFile, len(spanInfo.snapshot))
				fileCoverageInfo := CoverageMap[spanInfo.refFile]
				for idx, line := range spanInfo.snapshot {
					start := (*fileCoverageInfo.Pos)[3*idx]
					end := (*fileCoverageInfo.Pos)[3*idx+1]
					fmt.Fprintf(f, "Entry %d - [%d:%d] - %d\n", idx, start, end, line)
				}
				fmt.Fprintln(f, "=========================")
			}
		}

	}
	f.Close()
	return exp.wrappedExporter.ExportSpans(ctx, ss)
}

func (exp *CodecovExporter) Shutdown(ctx context.Context) error {
	return exp.wrappedExporter.Shutdown(ctx)
}

// newExporter returns a console exporter.
func myExporter(w io.Writer) (trace.SpanExporter, error) {
	wrapped, _ := stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
	return &CodecovExporter{wrappedExporter: wrapped}, nil

}

// newResource returns a resource describing this application.
func myResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fib"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}

func MyTraceProvider(f *os.File) (*trace.TracerProvider, error) {
	exp, err := myExporter(f)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(myResource()),
	)
	tp.RegisterSpanProcessor(getProcessorSingleton())
	return tp, nil
}
