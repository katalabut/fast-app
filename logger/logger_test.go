package logger

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestGlobalLoggerConcurrent(t *testing.T) {
	original := Logger()
	t.Cleanup(func() { SetLogger(original) })
	start := make(chan struct{})
	var wg sync.WaitGroup
	for n := 0; n < 8; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			for i := 0; i < 100; i++ {
				switch n % 4 {
				case 0:
					SetLogger(zap.NewNop().Sugar())
				case 1:
					_ = FromContext(context.Background())
				case 2:
					_, _ = InitLogger(defaultCfg, "concurrent")
				case 3:
					_ = New(zapcore.InfoLevel, defaultCfg)
				}
			}
		}(n)
	}
	close(start)
	wg.Wait()
}

func TestContextEnrichment(t *testing.T) {
	core, entries := observer.New(zap.InfoLevel)
	ctx := ToContext(context.Background(), zap.New(core).Sugar())
	ctx = ContextWithKV(ctx, "request_id", "first")
	base := ContextWithTags(ctx, "base")
	child := ContextWithTags(base, "child")
	child = ContextWithKV(child, "user", 1)
	Info(child, "child")
	Info(base, "base")
	logs := entries.All()
	tagFields := 0
	for _, field := range logs[0].Context {
		if field.Key == "tags" {
			tagFields++
		}
	}
	if tagFields != 1 {
		t.Fatalf("duplicate tag fields: %d", tagFields)
	}
	if logs[0].ContextMap()["request_id"] != "first" {
		t.Fatal("KV field absent")
	}
	if !reflect.DeepEqual(logs[0].ContextMap()["tags"], []interface{}{"base", "child"}) {
		t.Fatalf("tags missing: %+v", logs[0].ContextMap())
	}
	if !reflect.DeepEqual(logs[1].ContextMap()["tags"], []interface{}{"base"}) {
		t.Fatal("parent context tags changed")
	}
}
