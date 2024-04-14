package main

import (
	"embed"
	"errors"
	"gost/internal/core"
	"gost/internal/db"
	mainpage "gost/page"
	"log/slog"
	"net/http"
	"time"
)

//go:embed assets
var assets embed.FS

type Server struct {
	Store *db.MongoRepo
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", s.mainPage)

	mux.HandleFunc("POST /user", s.newUser)

	mux.HandleFunc("GET /messages", s.getMessages)
	mux.HandleFunc("POST /message", s.newMessage)

	mux.Handle("GET /assets/", http.FileServer(http.FS(assets)))

	return mux
}

func (s *Server) mainPage(w http.ResponseWriter, r *http.Request) {
	user, err := s.Store.UserGetByIP(r.Context(), r.RemoteAddr)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		slog.Error("getting user by IP", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := mainpage.Page(user != nil).Render(r.Context(), w); err != nil {
		slog.Error("rendering main page", "error", err)
	}
}

func (s *Server) newUser(w http.ResponseWriter, r *http.Request) {
	user := core.User{
		Username: r.PostFormValue("usernameBox"),
		IP:       r.RemoteAddr,
	}

	if err := s.Store.UserInsert(r.Context(), user); err != nil {
		slog.Error("storing new username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Swap user input with message input.
	if err := mainpage.MessageInput().Render(r.Context(), w); err != nil {
		slog.Error("rendering message input", "error", err)
	}
}

func (s *Server) newMessage(w http.ResponseWriter, r *http.Request) {
	user, err := s.Store.UserGetByIP(r.Context(), r.RemoteAddr) // Could be cached in server.
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			slog.Error("user not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		slog.Error("getting user by IP", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.Store.MessageInsert(r.Context(), core.Message{
		CreatedAt: time.Now(),
		Content:   r.PostFormValue("messageBox"),
		CreatedBy: user.Username,
	})
	if err != nil {
		slog.Error("inserting message", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := mainpage.MessageInput().Render(r.Context(), w); err != nil {
		slog.Error("rendering message input", "error", err)
	}
}

func (s *Server) getMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := s.Store.MessageSelectDesc(r.Context())
	if err != nil {
		slog.Error("getting messages", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := mainpage.MessageScreen(messages).Render(r.Context(), w); err != nil {
		slog.Error("rending message screen", "error", err)
	}
}
