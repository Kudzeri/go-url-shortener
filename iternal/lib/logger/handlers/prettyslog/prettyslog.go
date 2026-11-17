package slogpretty

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type Options struct {
	Level slog.Leveler
	Color bool
}

type output struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	writer io.Writer
}

type Handler struct {
	handler slog.Handler
	output  *output
	color   bool
}

var _ slog.Handler = (*Handler)(nil)

func New(writer io.Writer, opts Options) *Handler {
	out := &output{writer: writer}

	handler := slog.NewJSONHandler(&out.buffer, &slog.HandlerOptions{
		Level: opts.Level,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 {
				switch attr.Key {
				case slog.TimeKey, slog.LevelKey, slog.MessageKey:
					return slog.Attr{}
				}
			}

			return attr
		},
	})

	return &Handler{
		handler: handler,
		output:  out,
		color:   opts.Color,
	}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	h.output.mu.Lock()
	defer h.output.mu.Unlock()

	h.output.buffer.Reset()

	if err := h.handler.Handle(ctx, record); err != nil {
		return err
	}

	level := record.Level.String()
	message := record.Message

	if h.color {
		level = levelColor(record.Level) + level + "\033[0m"
		message = "\033[38;2;152;255;204m" + message + "\033[0m"
	}

	var line bytes.Buffer

	fmt.Fprintf(
		&line,
		"[%s] %s: %s\n",
		record.Time.Format("15:04:05.000"),
		level,
		message,
	)

	attributes := bytes.TrimSpace(h.output.buffer.Bytes())
	if !bytes.Equal(attributes, []byte("{}")) {
		var formatted bytes.Buffer
		if err := json.Indent(&formatted, attributes, "", "  "); err != nil {
			return err
		}

		if h.color {
			line.WriteString("\033[90m")
		}

		line.Write(formatted.Bytes())

		if h.color {
			line.WriteString("\033[0m")
		}

		line.WriteByte('\n')
	}

	_, err := h.output.writer.Write(line.Bytes())
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		handler: h.handler.WithAttrs(attrs),
		output:  h.output,
		color:   h.color,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		handler: h.handler.WithGroup(name),
		output:  h.output,
		color:   h.color,
	}
}

func levelColor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "\033[31m" // Красный.
	case level >= slog.LevelWarn:
		return "\033[33m" // Жёлтый.
	case level >= slog.LevelInfo:
		return "\033[32m" // Зелёный.
	default:
		return "\033[36m" // Голубой.
	}
}
