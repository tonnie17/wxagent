package web

import (
	"encoding/xml"
	"io"
	"log/slog"
)

func xmlParseRequest(body io.Reader, req interface{}) error {
	b, err := io.ReadAll(body)
	if err != nil {
		slog.Error("body read failed", slog.Any("err", err))
		return err
	}
	slog.Info("print body", slog.String("body", string(b)))

	if err := xml.Unmarshal(b, &req); err != nil {
		slog.Error("request parse failed", slog.Any("err", err))
		return err
	}

	return nil
}
