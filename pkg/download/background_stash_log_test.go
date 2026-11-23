package download

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
)

// sysErrEntry is a log.Entry that hands every SystemErr call to a channel.
type sysErrEntry struct {
	errs chan error
}

func (*sysErrEntry) Debugf(string, ...any) {}
func (*sysErrEntry) Infof(string, ...any)  {}
func (*sysErrEntry) Warnf(string, ...any)  {}
func (*sysErrEntry) Errorf(string, ...any) {}

func (e *sysErrEntry) WithFields(map[string]any) log.Entry { return e }

func (e *sysErrEntry) SystemErr(err error) { e.errs <- err }

type failingStasher struct{ err error }

func (f failingStasher) Stash(context.Context, string, string) (string, error) {
	return "", f.err
}

// TestBackgroundStashErrorSeverity ensures a failing async stash is logged, and
// that an expected outcome (module not found) is not logged as an error.
func TestBackgroundStashErrorSeverity(t *testing.T) {
	tests := []struct {
		name string
		mode mode.Mode
		err  error
		want slog.Level
	}{
		{"async not found", mode.Async, errors.E("stash", "no such module", errors.KindNotFound), slog.LevelInfo},
		{"async_redirect not found", mode.AsyncRedirect, errors.E("stash", "no such module", errors.KindNotFound), slog.LevelInfo},
		{"async unexpected", mode.Async, errors.E("stash", "disk on fire", errors.KindUnexpected), slog.LevelError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := mem.NewStorage()
			require.NoError(t, err)
			entry := &sysErrEntry{errs: make(chan error, 1)}
			dp := New(&Opts{
				Stasher:      failingStasher{tc.err},
				Storage:      s,
				DownloadFile: &mode.DownloadFile{Mode: tc.mode, DownloadURL: "https://gomods.io"},
			})
			ctx := log.SetEntryInContext(t.Context(), entry)
			_, _ = dp.Info(ctx, "github.com/athens-artifacts/happy-path", "v0.0.1")

			select {
			case got := <-entry.errs:
				require.Equal(t, tc.want, errors.Severity(got))
			case <-time.After(5 * time.Second):
				t.Fatal("background stash error was never logged")
			}
		})
	}
}
