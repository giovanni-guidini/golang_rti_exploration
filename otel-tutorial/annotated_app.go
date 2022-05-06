package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	// Uncomment to make coverage data work
	"fib/opentelem"
)

const name = "fib"

// App is a Fibonacci computation application.
type App struct {
	r io.Reader
	l *log.Logger
}

// NewApp returns a new App.
func NewApp(r io.Reader, l *log.Logger) *App {
	GoCover.Count[0]++
	return &App{r: r, l: l}
}

// Run starts polling users for Fibonacci number requests and writes results.
func (a *App) Run(ctx context.Context) error {
	GoCover.Count[1]++
	for {
		GoCover.Count[2]++
		// Each execution of the run loop, we should get a new "root" span and context.
		newCtx, span := otel.Tracer(name).Start(ctx, "Run")

		n, err := a.Poll(ctx)
		if err != nil {
			GoCover.Count[4]++
			return err
		}

		GoCover.Count[3]++
		a.Write(newCtx, n)
		span.End()
	}
}

// Poll asks a user for input and returns the request.
func (a *App) Poll(ctx context.Context) (uint, error) {
	GoCover.Count[5]++
	// Uncomment to make the coverage info work
	opentelem.RecordCoverageMap("/Users/giovannimguidini/Projects/Random/go-otel-tests/otel-tutorial/annotated_app.go", GoCover)

	_, span := otel.Tracer(name).Start(ctx, "Poll", trace.WithAttributes(attribute.Bool("coverage", true)))
	defer span.End()

	a.l.Print("What Fibonacci number would you like to know: ")

	var n uint
	_, err := fmt.Fscanf(a.r, "%d\n", &n)
	if err != nil {
		GoCover.Count[7]++
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	// Store n as a string to not overflow an int64.
	GoCover.Count[6]++
	nStr := strconv.FormatUint(uint64(n), 10)
	span.SetAttributes(attribute.String("request.n", nStr))

	return n, err
}

// Write writes the n-th Fibonacci number back to the user.
func (a *App) Write(ctx context.Context, n uint) {
	GoCover.Count[8]++
	var span trace.Span
	ctx, span = otel.Tracer(name).Start(ctx, "Write")
	defer span.End()

	f, err := func(ctx context.Context) (uint64, error) {
		GoCover.Count[10]++
		_, span := otel.Tracer(name).Start(ctx, "Fibonacci")
		defer span.End()
		f, err := Fibonacci(n)
		if err != nil {
			GoCover.Count[12]++
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		GoCover.Count[11]++
		return f, err
	}(ctx)
	GoCover.Count[9]++
	if err != nil {
		GoCover.Count[13]++
		a.l.Printf("Fibonacci(%d): %v\n", n, err)
	} else {
		GoCover.Count[14]++
		{
			a.l.Printf("Fibonacci(%d) = %d\n", n, f)
		}
	}
}

var GoCover = struct {
	Count   []uint32
	Pos     []uint32
	NumStmt []uint16
}{
	Count: make([]uint32, 15),
	Pos: []uint32{
		25, 27, 0x2002e, // [0]
		30, 31, 0x6002e, // [1]
		31, 36, 0x110006, // [2]
		40, 41, 0xd0003, // [3]
		36, 38, 0x40011, // [4]
		46, 58, 0x100037, // [5]
		65, 68, 0xf0002, // [6]
		58, 62, 0x30010, // [7]
		72, 77, 0x360032, // [8]
		87, 87, 0x100002, // [9]
		77, 81, 0x110036, // [10]
		85, 85, 0x100003, // [11]
		81, 84, 0x40011, // [12]
		87, 89, 0x30010, // [13]
		89, 91, 0x30008, // [14]
	},
	NumStmt: []uint16{
		1, // 0
		1, // 1
		3, // 2
		2, // 3
		1, // 4
		6, // 5
		3, // 6
		3, // 7
		4, // 8
		1, // 9
		4, // 10
		1, // 11
		2, // 12
		1, // 13
		1, // 14
	},
}
