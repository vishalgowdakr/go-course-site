package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

var (
	store       *sessions.CookieStore
	models      = make(map[string]*Model)
	modelsMutex sync.RWMutex
)

func init() {
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET environment variable is not set")
	}
	store = sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   true, // Set to true if using HTTPS
	}
}

func setModel(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		model, err := getOrCreateModel(c)
		if err != nil {
			return fmt.Errorf("failed to get or create model: %w", err)
		}
		c.Set("model", model)
		return next(c)
	}
}

func getOrCreateModel(c echo.Context) (*Model, error) {
	sessionID := getSessionID(c)
	modelsMutex.Lock()
	defer modelsMutex.Unlock()
	if model, exists := models[sessionID]; exists {
		return model, nil
	}
	log.Print("New session: ", sessionID)
	model := &Model{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	filep := filepath.Join(homeDir, "personal", "course-site", "lessons")
	if err := model.Init(filep, c); err != nil {
		return nil, fmt.Errorf("failed to initialize model: %w", err)
	}
	models[sessionID] = model
	return model, nil
}

func getSessionID(c echo.Context) string {
	sess, err := store.Get(c.Request(), "session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return uuid.New().String()
	}

	if sess.IsNew {
		sess.ID = uuid.New().String()
		if err = sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Failed to save session: %v", err)
		} else {
			log.Printf("New session saved: %s", sess.ID)
		}
	} else {
		log.Printf("Existing session found: %s", sess.ID)
	}

	return sess.ID
}
