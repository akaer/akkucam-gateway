// Package gateway implements the local HTTP API, request validation and error
// mapping. It contains no protocol or media logic.
package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"akkucam-gateway/internal/apierr"
	"akkucam-gateway/internal/discovery"
	"akkucam-gateway/internal/openapi"
	"akkucam-gateway/internal/ppcs"
	streammanager "akkucam-gateway/internal/stream"
	"akkucam-gateway/internal/version"
)

// Server owns the HTTP API and its injected dependencies.
type Server struct {
	webRoot       string
	ppcsMu        sync.Mutex
	streams       streamService
	probeFn       func(context.Context, string, string) (ppcs.ProbeResult, error)
	authFn        func(context.Context, string, string, string) (ppcs.AuthResult, error)
	channelsFn    func(context.Context, string, string, string) (ppcs.Inventory, error)
	motionFn      func(context.Context, string, string, string, int, bool) (ppcs.CameraChannel, error)
	waiDiscovery  func(context.Context) ([]discovery.Device, error)
	ppcsDiscovery func(context.Context) ([]discovery.Device, error)
	docs          http.Handler
}

// streamService is the stream-manager boundary used by the API.
type streamService interface {
	Root() string
	Start(context.Context, string, string, string, int, int) (streammanager.Result, error)
	Stop(string)
	Status(string) streammanager.Status
	ActiveCount() int
}

type discoverResponse struct {
	Devices    []discovery.Device `json:"devices"`
	DurationMs int64              `json:"durationMs"`
	Errors     map[string]string  `json:"errors,omitempty"`
}

type probeRequest struct {
	IP           string `json:"ip"`
	DID          string `json:"did"`
	SecurityCode string `json:"securityCode,omitempty"`
}

type motionRequest struct {
	IP           string `json:"ip"`
	DID          string `json:"did"`
	SecurityCode string `json:"securityCode"`
	Channel      int    `json:"channel"`
	Enabled      *bool  `json:"enabled"`
}

type liveRequest struct {
	IP           string `json:"ip"`
	DID          string `json:"did"`
	SecurityCode string `json:"securityCode,omitempty"`
	CameraIndex  int    `json:"cameraIndex,omitempty"`
	Quality      int    `json:"quality,omitempty"`
	ID           string `json:"id,omitempty"`
}

// New builds a server with the real protocol and discovery implementations.
func New(webRoot string, streams streamService) *Server {
	s := &Server{
		webRoot:       webRoot,
		streams:       streams,
		probeFn:       ppcs.Probe,
		authFn:        ppcs.Authenticate,
		channelsFn:    ppcs.QueryInventory,
		motionFn:      ppcs.SetMotion,
		waiDiscovery:  discovery.WAI,
		ppcsDiscovery: discovery.PPCS,
	}
	s.docs = http.StripPrefix("/api/docs/", http.FileServer(http.Dir(filepath.Join(webRoot, "swagger"))))
	return s
}

type route struct {
	method  string
	path    string
	handler http.HandlerFunc
}

func (s *Server) routeTable() []route {
	return []route{
		{http.MethodGet, "/api/health", s.health},
		{http.MethodGet, "/api/version", s.versionInfo},
		{http.MethodPost, "/api/discover", s.discover},
		{http.MethodPost, "/api/session/probe", s.probeSession},
		{http.MethodPost, "/api/session/authenticate", s.authenticateSession},
		{http.MethodPost, "/api/channels", s.channels},
		{http.MethodPost, "/api/motion", s.motion},
		{http.MethodPost, "/api/live/start", s.startLive},
		{http.MethodPost, "/api/live/stop", s.stopLive},
		{http.MethodGet, "/api/live/status", s.liveStatus},
		{http.MethodGet, "/api/openapi.json", openapi.Handler()},
		{http.MethodGet, "/api/docs", s.docsRedirect},
		{http.MethodGet, "/api/docs/", s.docsHandler},
	}
}

// Routes builds the HTTP handler tree. The route table is the single source of
// truth for API paths and methods.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	routes := s.routeTable()
	allowed := make(map[string]string)
	for _, r := range routes {
		mux.HandleFunc(r.method+" "+r.path, r.handler)
		if _, ok := allowed[r.path]; !ok {
			allowed[r.path] = r.method
		}
	}
	for path, method := range allowed {
		allowedMethod := method
		mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Allow", allowedMethod)
			apierr.Write(w, apierr.New(apierr.CodeMethodNotAllowed, http.StatusMethodNotAllowed, "method not allowed"))
		})
	}
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		apierr.Write(w, apierr.New(apierr.CodeNotFound, http.StatusNotFound, "API endpoint not found"))
	})
	mux.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir(s.streams.Root()))))
	mux.Handle("/", http.FileServer(http.Dir(s.webRoot)))
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"mode":   "local-only",
		"live":   "hls",
	})
}

func (s *Server) versionInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, struct {
		Version string `json:"version"`
		Commit  string `json:"commit,omitempty"`
	}{
		Version: version.Version,
		Commit:  version.Commit(),
	})
}

func (s *Server) discover(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), DiscoveryTimeout)
	defer cancel()

	type result struct {
		name    string
		devices []discovery.Device
		err     error
	}
	results := make(chan result, DiscoverySources)
	go func() {
		devices, err := s.waiDiscovery(ctx)
		results <- result{name: "wai", devices: devices, err: err}
	}()
	go func() {
		devices, err := s.ppcsDiscovery(ctx)
		results <- result{name: "ppcs", devices: devices, err: err}
	}()

	byIP := make(map[string]discovery.Device)
	errs := make(map[string]string)
	for range DiscoverySources {
		res := <-results
		if res.err != nil && !errors.Is(res.err, context.DeadlineExceeded) {
			errs[res.name] = res.err.Error()
		}
		for _, device := range res.devices {
			current := byIP[device.IP]
			current.IP = device.IP
			if len(device.DID) > len(current.DID) {
				current.DID = device.DID
			}
			if device.Model != "" {
				current.Model = device.Model
			}
			if device.PPCSPort != 0 {
				current.PPCSPort = device.PPCSPort
			}
			if current.MAC == "" {
				current.MAC = discovery.MACForIP(device.IP)
			}
			current.Local = true
			byIP[device.IP] = current
		}
	}

	devices := make([]discovery.Device, 0, len(byIP))
	for _, device := range byIP {
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].IP < devices[j].IP })

	response := discoverResponse{Devices: devices, DurationMs: time.Since(started).Milliseconds()}
	if len(errs) > 0 {
		response.Errors = errs
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) probeSession(w http.ResponseWriter, r *http.Request) {
	var request probeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apierr.Write(w, apierr.New(apierr.CodeInvalidJSON, http.StatusBadRequest, "invalid JSON request"))
		return
	}
	if net.ParseIP(request.IP) == nil && strings.TrimSpace(request.DID) == "" {
		apierr.Write(w, apierr.New(apierr.CodeValidation, http.StatusBadRequest, "ip or did is required"))
		return
	}
	s.ppcsMu.Lock()
	defer s.ppcsMu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), ProbeTimeout)
	defer cancel()
	result, err := s.probeFn(ctx, strings.TrimSpace(request.IP), strings.TrimSpace(request.DID))
	if err != nil {
		log.Printf("PPCS probe failed target=%s did=%s error=%v", request.IP, request.DID, err)
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) authenticateSession(w http.ResponseWriter, r *http.Request) {
	var request probeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apierr.Write(w, apierr.New(apierr.CodeInvalidJSON, http.StatusBadRequest, "invalid JSON request"))
		return
	}
	if net.ParseIP(request.IP) == nil && strings.TrimSpace(request.DID) == "" {
		apierr.Write(w, apierr.New(apierr.CodeValidation, http.StatusBadRequest, "ip or did is required"))
		return
	}
	s.ppcsMu.Lock()
	defer s.ppcsMu.Unlock()

	var result ppcs.AuthResult
	var err error
	for attempt := 1; attempt <= AuthAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(r.Context(), AuthTimeout)
		result, err = s.authFn(attemptCtx, strings.TrimSpace(request.IP), strings.TrimSpace(request.DID), request.SecurityCode)
		cancel()
		if err == nil || errors.Is(err, ppcs.ErrAuthRejected) {
			break
		}
		if attempt < AuthAttempts {
			time.Sleep(AuthRetryDelay)
		}
	}
	request.SecurityCode = ""
	if err != nil {
		log.Printf("PPCS authentication failed target=%s did=%s error=%v", request.IP, request.DID, err)
		writeError(w, err)
		return
	}
	log.Printf("PPCS authentication succeeded target=%s did=%s duration=%s", request.IP, request.DID, result.RoundTrip)
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) channels(w http.ResponseWriter, r *http.Request) {
	var request probeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apierr.Write(w, apierr.New(apierr.CodeInvalidJSON, http.StatusBadRequest, "invalid JSON request"))
		return
	}
	if net.ParseIP(request.IP) == nil || strings.TrimSpace(request.DID) == "" || request.SecurityCode == "" {
		apierr.Write(w, apierr.New(apierr.CodeValidation, http.StatusBadRequest, "valid ip, did and securityCode are required"))
		return
	}
	s.ppcsMu.Lock()
	defer s.ppcsMu.Unlock()
	if s.streams.ActiveCount() > 0 {
		apierr.Write(w, apierr.New(apierr.CodeConflict, http.StatusConflict, "stop active streams before reloading camera inventory"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), SessionTimeout)
	defer cancel()
	inventory, err := s.channelsFn(ctx, strings.TrimSpace(request.IP), strings.TrimSpace(request.DID), request.SecurityCode)
	request.SecurityCode = ""
	if err != nil {
		log.Printf("PPCS inventory failed target=%s did=%s error=%v", request.IP, request.DID, err)
		writeError(w, err)
		return
	}
	if inventory.Hub.MAC == "" {
		inventory.Hub.MAC = discovery.MACForIP(strings.TrimSpace(request.IP))
	}
	writeJSON(w, http.StatusOK, map[string]any{"did": strings.TrimSpace(request.DID), "hub": inventory.Hub, "channels": inventory.Channels})
}

func (s *Server) motion(w http.ResponseWriter, r *http.Request) {
	var request motionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apierr.Write(w, apierr.New(apierr.CodeInvalidJSON, http.StatusBadRequest, "invalid JSON request"))
		return
	}
	if net.ParseIP(request.IP) == nil || strings.TrimSpace(request.DID) == "" || request.SecurityCode == "" || request.Channel < 0 || request.Channel > ppcs.MaxMotionChannel || request.Enabled == nil {
		apierr.Write(w, apierr.New(apierr.CodeValidation, http.StatusBadRequest, "valid ip, did, securityCode, channel and enabled are required"))
		return
	}
	s.ppcsMu.Lock()
	defer s.ppcsMu.Unlock()
	if s.streams.ActiveCount() > 0 {
		apierr.Write(w, apierr.New(apierr.CodeConflict, http.StatusConflict, "stop active streams before changing motion detection"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), SessionTimeout)
	defer cancel()
	channel, err := s.motionFn(ctx, strings.TrimSpace(request.IP), strings.TrimSpace(request.DID), request.SecurityCode, request.Channel, *request.Enabled)
	request.SecurityCode = ""
	if err != nil {
		log.Printf("PPCS motion failed target=%s did=%s channel=%d error=%v", request.IP, request.DID, request.Channel, err)
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channel)
}

func (s *Server) startLive(w http.ResponseWriter, r *http.Request) {
	var request liveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apierr.Write(w, apierr.New(apierr.CodeInvalidJSON, http.StatusBadRequest, "invalid JSON request"))
		return
	}
	if net.ParseIP(request.IP) == nil || strings.TrimSpace(request.DID) == "" || request.CameraIndex < 0 || request.CameraIndex > MaxCameraIndex || request.Quality < 0 || request.Quality > MaxQuality {
		apierr.Write(w, apierr.New(apierr.CodeValidation, http.StatusBadRequest, "valid ip, did, cameraIndex 0-3 and quality 0-4 are required"))
		return
	}
	s.ppcsMu.Lock()
	result, err := s.streams.Start(r.Context(), strings.TrimSpace(request.IP), strings.TrimSpace(request.DID), request.SecurityCode, request.CameraIndex, request.Quality)
	s.ppcsMu.Unlock()
	request.SecurityCode = ""
	if err != nil {
		log.Printf("start live failed target=%s did=%s camera=%d quality=%d error=%v", request.IP, request.DID, request.CameraIndex, request.Quality, err)
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) stopLive(w http.ResponseWriter, r *http.Request) {
	var request liveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ID == "" {
		apierr.Write(w, apierr.New(apierr.CodeValidation, http.StatusBadRequest, "stream id is required"))
		return
	}
	s.streams.Stop(request.ID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) liveStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		apierr.Write(w, apierr.New(apierr.CodeValidation, http.StatusBadRequest, "stream id is required"))
		return
	}
	writeJSON(w, http.StatusOK, s.streams.Status(id))
}

func (s *Server) docsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/docs/", http.StatusMovedPermanently)
}

func (s *Server) docsHandler(w http.ResponseWriter, r *http.Request) {
	s.docs.ServeHTTP(w, r)
}

func writeError(w http.ResponseWriter, err error) {
	var apiErr *apierr.Error
	if errors.As(err, &apiErr) {
		apierr.Write(w, apiErr)
		return
	}
	switch {
	case errors.Is(err, ppcs.ErrAuthRejected):
		apierr.Write(w, apierr.New(apierr.CodeUnauthorized, http.StatusUnauthorized, "security code rejected"))
	case errors.Is(err, ppcs.ErrAuthTimeout), errors.Is(err, ppcs.ErrLanTimeout), errors.Is(err, ppcs.ErrP2PReadyTimeout), errors.Is(err, context.DeadlineExceeded):
		apierr.Write(w, apierr.New(apierr.CodeUpstreamTimeout, http.StatusGatewayTimeout, "device did not respond in time"))
	case errors.Is(err, ppcs.ErrPeerClosed):
		apierr.Write(w, apierr.New(apierr.CodeUpstreamError, http.StatusBadGateway, "device closed the session"))
	case errors.Is(err, streammanager.ErrFFmpegStart):
		apierr.Write(w, apierr.New(apierr.CodeInternal, http.StatusInternalServerError, "could not start the media pipeline"))
	default:
		// Anything that is not a known device error is an internal failure
		// (ADR-0016). apierr.Write emits internal_error/500 with a generic
		// message; the caller has already logged the details.
		apierr.Write(w, err)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}

// RequestLogger logs method, path, remote host and duration without logging
// request bodies or credentials.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		//nolint:gosec // G706: path masked with %q, host from SplitHostPort; log interprets no escapes (ADR-0033)
		log.Printf("%s %q from=%s duration=%s", r.Method, r.URL.Path, host, time.Since(started).Round(time.Millisecond))
	})
}
