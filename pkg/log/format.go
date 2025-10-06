package log

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/fatih/color"
)

func getGCPFormatter(level slog.Level, w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.LevelKey:
				return slog.String("severity", a.Value.String())
			case slog.MessageKey:
				return slog.String("message", a.Value.String())
			case slog.TimeKey:
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
			default:
				return a
			}
		},
	}))
}

const lightGrey = 0xffccc

func getDevFormatter(level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format(time.Kitchen))
			}
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				var colored string
				switch level {
				case slog.LevelDebug:
					colored = color.New(lightGrey).Sprint(level)
				case slog.LevelWarn:
					colored = color.YellowString(level.String())
				case slog.LevelError:
					colored = color.RedString(level.String())
				default:
					colored = color.CyanString(level.String())
				}
				return slog.String(slog.LevelKey, colored)
			}
			if len(groups) == 0 {
				return slog.Attr{
					Key:   color.MagentaString(a.Key),
					Value: a.Value,
				}
			}
			return a
		},
	}))
}

func parseFormat(format string, level slog.Level, w io.Writer) *slog.Logger {
	if format == "json" {
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
	}

	return getDevFormatter(level)
}
