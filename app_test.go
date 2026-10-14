package fastapp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Start owns process termination, so exercise it in a child test process.
func TestAppLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name   string
		signal os.Signal
		code   int
	}{
		{"sigterm", syscall.SIGTERM, 0},
		{"interrupt", os.Interrupt, 0},
		{"cancel", nil, 0},
		{"returns-nil", nil, 0},
		{"already-cancelled", nil, 0},
		{"run-error", nil, 1},
		{"panic", nil, 1},
		{"shutdown-error", nil, 1},
		{"timeout", nil, 1},
		{"long-shutdown", nil, 0},
		{"watchdog", nil, 1},
		{"empty", nil, 0},
		{"max-timeout", nil, 0},
		{"disabled-error", nil, 1},
		{"disabled-panic", nil, 1},
		{"disabled-shutdown-error", nil, 1},
		{"shutdown-panic", nil, 1},
		{"watchdog-blocked-sync", nil, 1},
		{"success-blocked-sync", nil, 1},
		{"error-blocked-sync", nil, 1},
		{"watchdog-blocked-write", nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.signal != nil && runtime.GOOS == "windows" {
				t.Skip("process signals are not supported on Windows")
			}
			childTimeout := 25 * time.Second
			if strings.Contains(tc.name, "blocked") {
				childTimeout = 8 * time.Second
			}
			ctx, cancel := context.WithTimeout(context.Background(), childTimeout)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAppLifecycleProcess$")
			cmd.Env = append(os.Environ(), "FASTAPP_LIFECYCLE_TEST="+tc.name)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			var stderr strings.Builder
			cmd.Stderr = &stderr
			startedAt := time.Now()
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = cmd.Process.Kill() })
			scanner := bufio.NewScanner(stdout)
			var output strings.Builder
			ready := false
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Fprintln(&output, line)
				if line == "ready" {
					ready = true
					if tc.signal != nil {
						if err := cmd.Process.Signal(tc.signal); err != nil {
							t.Fatal(err)
						}
					} else if tc.name != "empty" && tc.name != "already-cancelled" && tc.name != "max-timeout" {
						if _, err := io.WriteString(stdin, "stop\n"); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			if err := scanner.Err(); err != nil {
				t.Fatal(err)
			}
			err = cmd.Wait()
			if (strings.Contains(tc.name, "blocked") || tc.name == "watchdog") && time.Since(startedAt) < 5*time.Second {
				t.Errorf("watchdog exited before the shutdown grace period: %s", output.String())
			}
			if ctx.Err() != nil {
				t.Fatalf("child timed out: %s\n%s", output.String(), stderr.String())
			}
			code := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				code = exitErr.ExitCode()
			}
			if code != tc.code {
				t.Errorf("exit code = %d, want %d: %s\n%s", code, tc.code, output.String(), stderr.String())
			}
			if !ready {
				t.Error("child never became ready")
			}
			want := []string{"synced"}
			if strings.HasPrefix(tc.name, "disabled-") {
				want = append(want, "peer-survived")
			}
			if strings.HasSuffix(tc.name, "blocked-sync") {
				want = []string{"sync-enter"}
			}
			if tc.name == "watchdog-blocked-write" {
				want = []string{"write-enter"}
			}
			if tc.name != "empty" {
				want = append(want, "peer-shutdown", "peer-returned")
				if !strings.HasPrefix(tc.name, "watchdog") {
					want = append(want, "primary-shutdown")
				}
			}
			if strings.HasPrefix(tc.name, "watchdog") && tc.name != "watchdog-blocked-write" && !strings.Contains(output.String(), "Graceful shutdown watchdog triggered") {
				t.Errorf("watchdog did not trigger: %s", output.String())
			}
			for _, marker := range want {
				if !strings.Contains(output.String(), marker+"\n") {
					t.Errorf("missing %q: %s", marker, output.String())
				}
			}
		})
	}
}

type lifecycleService struct {
	run      func(context.Context) error
	shutdown func(context.Context) error
}

func (s lifecycleService) Run(ctx context.Context) error      { return s.run(ctx) }
func (s lifecycleService) Shutdown(ctx context.Context) error { return s.shutdown(ctx) }

type lifecycleSink struct{ mode string }

func (s lifecycleSink) Write(p []byte) (int, error) {
	if s.mode == "watchdog-blocked-write" && strings.Contains(string(p), "Waiting for application shutdown") {
		fmt.Println("write-enter")
		select {}
	}
	return os.Stdout.Write(p)
}
func (s lifecycleSink) Sync() error {
	if strings.HasSuffix(s.mode, "blocked-sync") {
		fmt.Println("sync-enter")
		select {}
	}
	fmt.Println("synced")
	return nil
}

func TestAppLifecycleProcess(t *testing.T) {
	mode := os.Getenv("FASTAPP_LIFECYCLE_TEST")
	if mode == "" {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), lifecycleSink{mode: mode}, zap.InfoLevel)
	shutdownTimeout := time.Second
	if mode == "timeout" || strings.HasPrefix(mode, "watchdog") || strings.HasSuffix(mode, "blocked-sync") {
		shutdownTimeout = 20 * time.Millisecond
	}
	if mode == "long-shutdown" {
		shutdownTimeout = 12 * time.Second
	}
	if mode == "max-timeout" {
		shutdownTimeout = time.Duration(1<<63 - 1)
	}
	cfg := Config{}
	cfg.Logger.Level = "info"
	a := New(cfg, WithContext(ctx), WithShutdownTimeout(shutdownTimeout))
	a.logger = zap.New(core).Sugar()
	if strings.HasPrefix(mode, "disabled-") {
		a.opts.stopAllOnErr = false
	}
	if mode == "empty" {
		cancel()
		fmt.Println("ready")
		a.Start()
		t.Fatal("Start returned")
	}
	if mode == "already-cancelled" || mode == "max-timeout" {
		cancel()
	}
	peerStarted := make(chan struct{})
	peerStopped := make(chan struct{})
	primaryFailed := make(chan struct{})
	primary := lifecycleService{
		run: func(ctx context.Context) error {
			<-peerStarted
			fmt.Println("ready")
			if mode == "sigterm" || mode == "interrupt" || mode == "already-cancelled" || mode == "max-timeout" {
				<-ctx.Done()
				return ctx.Err()
			}
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			switch mode {
			case "run-error", "error-blocked-sync", "disabled-error", "disabled-shutdown-error":
				if strings.HasPrefix(mode, "disabled-") {
					close(primaryFailed)
				}
				return errors.New("run failed")
			case "panic", "disabled-panic":
				if mode == "disabled-panic" {
					close(primaryFailed)
				}
				panic("run panicked")
			case "returns-nil":
				cancel()
				return nil
			default:
				cancel()
				<-ctx.Done()
				return ctx.Err()
			}
		},
		shutdown: func(ctx context.Context) error {
			switch mode {
			case "shutdown-error", "disabled-shutdown-error":
				fmt.Println("primary-shutdown")
				return errors.New("shutdown failed")
			case "shutdown-panic":
				fmt.Println("primary-shutdown")
				panic("shutdown panicked")
			case "timeout":
				<-ctx.Done()
				fmt.Println("primary-shutdown")
				return ctx.Err()
			case "long-shutdown":
				time.Sleep(11 * time.Second)
			case "max-timeout":
				time.Sleep(100 * time.Millisecond)
			case "watchdog", "watchdog-blocked-sync", "watchdog-blocked-write":
				select {} // Deliberately violate the shutdown context contract.
			}
			fmt.Println("primary-shutdown")
			return nil
		},
	}
	peer := lifecycleService{
		run: func(ctx context.Context) error {
			close(peerStarted)
			if strings.HasPrefix(mode, "disabled-") {
				<-primaryFailed
				select {
				case <-ctx.Done():
				case <-time.After(200 * time.Millisecond):
					fmt.Println("peer-survived")
					cancel()
				}
			}
			<-ctx.Done()
			<-peerStopped
			fmt.Println("peer-returned")
			return ctx.Err()
		},
		shutdown: func(ctx context.Context) error {
			time.Sleep(20 * time.Millisecond)
			fmt.Println("peer-shutdown")
			close(peerStopped)
			return nil
		},
	}
	a.Add(primary).Add(peer)
	a.Start()
	t.Fatal("Start returned")
}

func TestGracefulShutdown(t *testing.T) {
	failure := errors.New("cleanup failed")
	for _, tc := range []struct {
		name    string
		failure error
	}{{"success", nil}, {"error", failure}} {
		t.Run(tc.name, func(t *testing.T) {
			a := &App{opts: options{shutdownTimeout: time.Second}}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			var calls []int
			err := a.GracefulShutdown(ctx,
				func(ctx context.Context) error {
					if ctx.Err() != nil {
						t.Error("shutdown context already cancelled")
					}
					if _, ok := ctx.Deadline(); !ok {
						t.Error("shutdown context has no deadline")
					}
					calls = append(calls, 1)
					return tc.failure
				},
				func(context.Context) error { calls = append(calls, 2); return nil },
			)()
			if !errors.Is(err, tc.failure) {
				t.Errorf("error = %v, want %v", err, tc.failure)
			}
			wantCalls := "[1 2]"
			if tc.failure != nil {
				wantCalls = "[1]" // Preserve the existing first-error return contract.
			}
			if fmt.Sprint(calls) != wantCalls {
				t.Errorf("cleanup calls = %v, want %s", calls, wantCalls)
			}
		})
	}
}

func TestWatchdogDelay(t *testing.T) {
	const maxDuration = time.Duration(1<<63 - 1)
	for _, tc := range []struct {
		name    string
		timeout time.Duration
		want    time.Duration
	}{
		{"zero", 0, watchdogGracePeriod},
		{"default", defaultShutdownTimeout, 10 * time.Second},
		{"below boundary", maxDuration - watchdogGracePeriod - 1, maxDuration - 1},
		{"boundary", maxDuration - watchdogGracePeriod, maxDuration},
		{"above boundary", maxDuration - watchdogGracePeriod + 1, maxDuration},
		{"maximum", maxDuration, maxDuration},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := watchdogDelay(tc.timeout); got != tc.want {
				t.Errorf("watchdogDelay(%v) = %v, want %v", tc.timeout, got, tc.want)
			}
		})
	}
}
