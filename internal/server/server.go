package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"

	cfg "github.com/gasparian/go-api-template/internal/config"
	strg "github.com/gasparian/go-api-template/pkg/storage"
)

const visitorCookieName = "visitor_id"

// App holds the application dependencies
type App struct {
	Router  *mux.Router
	Logger  *zap.Logger
	Version string
	Name    string
	storage strg.Storage
	corsCfg cfg.CORSConfig
}

// Initialize sets up the application
func (a *App) Initialize(config cfg.ApplicationConfig, corsCfg cfg.CORSConfig, storage strg.Storage) {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	a.Logger = logger
	a.Router = mux.NewRouter()
	a.Version = config.Version
	a.Name = config.Name
	a.storage = storage
	a.corsCfg = corsCfg
	a.Router.Use(a.loggingMiddleware)
	a.initializeRoutes()
}

// initializeRoutes sets up the application routes
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/internal/ping", a.handlePing).Methods("GET")
	a.Router.HandleFunc("/internal/version", a.handleVersion).Methods("GET")

	allowedOrigins := a.corsCfg.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost", "http://localhost:*"}
	}
	allowedMethods := a.corsCfg.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodOptions}
	}
	allowCredentials := true
	if a.corsCfg.AllowCredentials != nil {
		allowCredentials = *a.corsCfg.AllowCredentials
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   allowedMethods,
		AllowCredentials: allowCredentials,
	})
	a.Router.Handle("/api/v1/users", c.Handler(http.HandlerFunc(a.handleUsersGet))).Methods("GET")
	a.Router.Handle("/api/v1/users", c.Handler(http.HandlerFunc(a.handleUsersPost))).Methods("POST", "OPTIONS")
}

// loggingMiddleware logs the Referer header using Zap
func (a *App) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// NOTE: don't spoil logs with lots of pings
		if path == "/ping" {
			next.ServeHTTP(w, r)
			return
		}
		referer := r.Referer()
		a.Logger.Info("Incoming request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("referer", referer),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Time("timestamp", time.Now()),
		)
		next.ServeHTTP(w, r)
	})
}

// handlePing responds with "pong"
func (a *App) handlePing(w http.ResponseWriter, r *http.Request) {
	if a.storage.Ping() == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("can't reach storage"))
	}
}

// handleVersion responds with the application version
func (a *App) handleVersion(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"version": a.Version}
	a.respondWithJSON(w, http.StatusOK, response)
}

// getVisitorID extracts the visitor ID from the request cookies.
func (a *App) getVisitorID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(visitorCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// No cookie found; return empty string without error
			return "", nil
		}
		// Log unexpected errors and return them
		a.Logger.Error("error reading cookie", zap.Error(err))
		return "", err
	}
	return cookie.Value, nil
}

// handleUsersPost ...
func (a *App) handleUsersPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return // just return, and middleware will add needed headers
	}
	visitorID, err := a.getVisitorID(r)
	if errors.Is(err, http.ErrNoCookie) {
		a.Logger.Error("error reading cookie", zap.Error(err))
		a.respondWithError(w, http.StatusInternalServerError, "error reading cookie")
		return
	} else if err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if r.Body == nil {
		a.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var data PostRequestModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = a.storage.Set(data.UserID, data.ItemID, visitorID)

	var status int
	if err != nil {
		a.Logger.Error(fmt.Sprintf("failed to update user %s", data.UserID), zap.Error(err))
		status = http.StatusInternalServerError
	} else {
		status = http.StatusNoContent
	}

	w.WriteHeader(status)

	a.Logger.Debug("Received data",
		zap.String("user_id", data.UserID),
		zap.String("item_id", data.ItemID),
	)
}

// handleUsersGet ...
func (a *App) handleUsersGet(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		a.respondWithError(w, http.StatusBadRequest, "Missing user_id parameter")
		return
	}

	userData, err := a.storage.Get(userID)
	if err != nil {
		a.Logger.Error(fmt.Sprintf("failed to retrieve user %s", userID), zap.Error(err))
		a.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		a.Logger.Error("failed to encode user data", zap.Error(err))
		a.respondWithError(w, http.StatusInternalServerError, "Failed to encode response")
	}

	a.Logger.Debug("Retrieved user data", zap.String("user_id", userID))
}

// respondWithError sends an error response in JSON format
func (a *App) respondWithError(w http.ResponseWriter, code int, message string) {
	response := map[string]string{"error": message}
	a.respondWithJSON(w, code, response)
}

// respondWithJSON sends a response in JSON format
func (a *App) respondWithJSON(w http.ResponseWriter, code int, payload map[string]string) {
	response, err := json.Marshal(payload)
	if err != nil {
		a.Logger.Error("Failed to marshal JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Run starts the HTTP server
func (a *App) Run(config cfg.ServerConfig) {
	timeoutSecDuration := time.Duration(config.Timeout) * time.Second
	srv := &http.Server{
		Handler:      a.Router,
		Addr:         config.Addr,
		WriteTimeout: timeoutSecDuration,
		ReadTimeout:  timeoutSecDuration,
	}
	a.Logger.Info("Starting server", zap.String("address", config.Addr))
	if err := srv.ListenAndServe(); err != nil {
		a.Logger.Fatal("Failed to start server", zap.Error(err))
		os.Exit(1)
	}
}
