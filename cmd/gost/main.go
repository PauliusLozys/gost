package main

import (
	"gost/internal/db"
	"log/slog"
	"net/http"
)

func main() {

	db, cleanup, err := db.New()
	if err != nil {
		slog.Error("loading DB", "error", err)
		return
	}
	defer cleanup()

	s := Server{
		Store: db,
	}

	slog.Info("running")
	err = http.ListenAndServe(":8080", s.Router())
	if err != nil {
		slog.Error("running server", "error", err)
	}
}
