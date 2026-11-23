package restful

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	restBasePath  = "/api/v1"
	serverVersion = "0.0.6"
	maxJSONBody   = 1 << 20
)

type apiErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type apiErrorResponse struct {
	Error apiErrorPayload `json:"error"`
}

func (inst *innerWebServer) handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(restBasePath+"/health", inst.handleHealth)
	mux.HandleFunc(restBasePath+"/status", inst.handleStatus)
	mux.HandleFunc(restBasePath+"/engine", inst.handleEngine)
	mux.HandleFunc(restBasePath+"/engine/start", inst.handleStartEngine)
	mux.HandleFunc(restBasePath+"/engine/stop", inst.handleStopEngine)
	mux.HandleFunc(restBasePath+"/engine/restart", inst.handleRestartEngine)
	mux.HandleFunc(restBasePath+"/transport", inst.handleTransport)
	mux.HandleFunc(restBasePath+"/transport/play", inst.handlePlay)
	mux.HandleFunc(restBasePath+"/transport/record", inst.handleRecord)
	mux.HandleFunc(restBasePath+"/transport/pause", inst.handlePause)
	mux.HandleFunc(restBasePath+"/transport/resume", inst.handleResume)
	mux.HandleFunc(restBasePath+"/transport/stop", inst.handleTransportStop)
	mux.HandleFunc(restBasePath+"/transport/position", inst.handleSeek)
	mux.HandleFunc(restBasePath+"/cube", inst.handleCube)
	mux.HandleFunc(restBasePath+"/banks", inst.handleBanks)
	mux.HandleFunc(restBasePath+"/banks/{bankIndex}", inst.handleBank)
	mux.HandleFunc(restBasePath+"/steps", inst.handleSteps)
	mux.HandleFunc(restBasePath+"/steps/{stepIndex}", inst.handleStep)
	mux.HandleFunc(restBasePath+"/tracks", inst.handleTracks)
	mux.HandleFunc(restBasePath+"/tracks/{trackIndex}", inst.handleTrack)
	mux.HandleFunc(restBasePath+"/cells", inst.handleCells)
	mux.HandleFunc(restBasePath+"/cells/{cellName}", inst.handleCell)
	mux.HandleFunc(restBasePath+"/actions", inst.handleActions)
	mux.HandleFunc("/", inst.handleSPA)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if inst.context != nil && inst.context.Logger != nil {
					inst.context.Logger.Error("REST server panic: %v", recovered)
				}
				writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
			}
		}()
		mux.ServeHTTP(w, r)
	})
}

func (inst *innerWebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"version": serverVersion,
	})
}

func (inst *innerWebServer) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "endpoint not found", map[string]any{
		"path": r.URL.Path,
	})
}

func allowMethods(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, method := range methods {
		if r.Method == method {
			return true
		}
		if r.Method == http.MethodHead && method == http.MethodGet {
			return true
		}
	}

	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", map[string]any{
		"method": r.Method,
	})
	return false
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if status == http.StatusNoContent || value == nil {
		return
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, apiErrorResponse{
		Error: apiErrorPayload{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is required")
	}

	body := http.MaxBytesReader(w, r.Body, maxJSONBody)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain one JSON object")
	}
	return nil
}

func parsePathIndex(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	text := r.PathValue(name)
	value, err := strconv.Atoi(text)
	if err != nil || value < 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PATH_PARAMETER",
			fmt.Sprintf("%s must be a non-negative integer", name),
			map[string]any{"field": name, "value": text})
		return 0, false
	}
	return value, true
}
