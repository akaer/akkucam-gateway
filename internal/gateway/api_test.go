package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"akkucam-gateway/internal/discovery"
	"akkucam-gateway/internal/ppcs"
	streammanager "akkucam-gateway/internal/stream"
	"akkucam-gateway/internal/version"
)

type fakeStreams struct {
	startRequest struct {
		ip, did, code    string
		channel, quality int
	}
	startResult streammanager.Result
	startErr    error
	stopped     []string
	status      streammanager.Status
	root        string
	active      int
}

func (f *fakeStreams) Root() string { return f.root }
func (f *fakeStreams) Start(_ context.Context, ip, did, code string, channel, quality int) (streammanager.Result, error) {
	f.startRequest.ip, f.startRequest.did, f.startRequest.code = ip, did, code
	f.startRequest.channel, f.startRequest.quality = channel, quality
	return f.startResult, f.startErr
}
func (f *fakeStreams) Stop(id string)                     { f.stopped = append(f.stopped, id) }
func (f *fakeStreams) Status(string) streammanager.Status { return f.status }
func (f *fakeStreams) ActiveCount() int                   { return f.active }

func testServer(t *testing.T) (*Server, http.Handler, *fakeStreams) {
	t.Helper()
	streams := &fakeStreams{root: t.TempDir()}
	webRoot := t.TempDir()
	swaggerDir := filepath.Join(webRoot, "swagger")
	if err := os.MkdirAll(swaggerDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(swaggerDir, "index.html"), []byte("<!doctype html><title>swagger</title>"), 0o640); err != nil {
		t.Fatal(err)
	}
	s := New(webRoot, streams)
	s.probeFn = func(context.Context, string, string) (ppcs.ProbeResult, error) {
		return ppcs.ProbeResult{Ready: true}, nil
	}
	s.authFn = func(context.Context, string, string, string) (ppcs.AuthResult, error) {
		return ppcs.AuthResult{Authenticated: true}, nil
	}
	s.channelsFn = func(context.Context, string, string, string) (ppcs.Inventory, error) { return ppcs.Inventory{}, nil }
	s.motionFn = func(context.Context, string, string, string, int, bool) (ppcs.CameraChannel, error) {
		return ppcs.CameraChannel{}, nil
	}
	s.waiDiscovery = func(context.Context) ([]discovery.Device, error) { return nil, nil }
	s.ppcsDiscovery = func(context.Context) ([]discovery.Device, error) { return nil, nil }
	return s, s.Routes(), streams
}

func request(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func decodeError(t *testing.T, response *httptest.ResponseRecorder) (string, string) {
	t.Helper()
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Error == "" {
		t.Fatal("error message is empty")
	}
	return body.Error, body.Code
}

func TestChannelsContract(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	s.channelsFn = func(context.Context, string, string, string) (ppcs.Inventory, error) {
		return ppcs.Inventory{Hub: ppcs.HubInfo{Firmware: "1.0.0.12"}, Channels: []ppcs.CameraChannel{
			{Index: 0, Name: "Einfahrt", Model: "SAPP-TAS"},
			{Index: 1, Name: "Garten", Model: "RVDP-TAS"},
		}}, nil
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/channels", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code"}`)

	// Assert
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"name":"Garten"`) {
		t.Fatalf("body=%s", response.Body.String())
	}
}

func TestChannelsDoesNotEchoSecret(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	secret := "session-only-secret"
	s.channelsFn = func(context.Context, string, string, string) (ppcs.Inventory, error) {
		return ppcs.Inventory{}, nil
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/channels", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"`+secret+`"}`)

	// Assert
	if strings.Contains(response.Body.String(), secret) {
		t.Fatal("channel response leaked security code")
	}
}

func TestChannelsPassesSecurityCodeToSession(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	secret := "session-only-secret"
	seen := ""
	s.channelsFn = func(_ context.Context, _, _, code string) (ppcs.Inventory, error) {
		seen = code
		return ppcs.Inventory{}, nil
	}

	// Act
	request(t, handler, http.MethodPost, "/api/channels", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"`+secret+`"}`)

	// Assert
	if seen != secret {
		t.Fatalf("session received %q, want the request secret", seen)
	}
}

func TestInventoryRejectsWhileStreamsAreActive(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)
	streams.active = 1

	// Act
	response := request(t, handler, http.MethodPost, "/api/channels", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code"}`)

	// Assert
	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d, want 409", response.Code)
	}
}

func TestMotionRejectsWhileStreamsAreActive(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)
	streams.active = 1

	// Act
	response := request(t, handler, http.MethodPost, "/api/motion", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code","channel":0,"enabled":true}`)

	// Assert
	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d, want 409", response.Code)
	}
}

func TestMotionContractSupportsExplicitFalse(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	var gotChannel int
	var gotEnabled bool
	called := false
	s.motionFn = func(_ context.Context, _, _, _ string, channel int, enabled bool) (ppcs.CameraChannel, error) {
		called = true
		gotChannel, gotEnabled = channel, enabled
		return ppcs.CameraChannel{Index: channel, MotionEnabled: &enabled}, nil
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/motion", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code","channel":1,"enabled":false}`)

	// Assert
	if response.Code != http.StatusOK || !called {
		t.Fatalf("status=%d called=%v", response.Code, called)
	}
	if gotChannel != 1 || gotEnabled {
		t.Fatalf("channel=%d enabled=%v", gotChannel, gotEnabled)
	}
}

func TestDiscoveryMergesSourcesWithoutTruncatingDID(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	s.waiDiscovery = func(context.Context) ([]discovery.Device, error) {
		return []discovery.Device{{IP: "192.0.2.10", DID: "CMAG-000001-ABCDE", Model: "SMH-T1R", Local: true}}, nil
	}
	s.ppcsDiscovery = func(context.Context) ([]discovery.Device, error) {
		return []discovery.Device{{IP: "192.0.2.10", DID: "CMAG", PPCSPort: 17500, Local: true}}, nil
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/discover", "")

	// Assert
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
	var body discoverResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Devices) != 1 {
		t.Fatalf("devices=%+v", body.Devices)
	}
	device := body.Devices[0]
	if device.DID != "CMAG-000001-ABCDE" || device.Model != "SMH-T1R" || device.PPCSPort != 17500 {
		t.Fatalf("merged device=%+v", device)
	}
}

func TestHealthContract(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)

	// Act
	response := request(t, handler, http.MethodGet, "/api/health", "")

	// Assert
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("content type = %q", got)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" || body["mode"] != "local-only" || body["live"] != "hls" {
		t.Fatalf("unexpected health body: %#v", body)
	}
}

func TestVersionContract(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)

	// Act
	response := request(t, handler, http.MethodGet, "/api/version", "")

	// Assert
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	var body struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Version != version.Version {
		t.Fatalf("version = %q, want %q", body.Version, version.Version)
	}
	if body.Commit != version.Commit() {
		t.Fatalf("commit = %q, want %q", body.Commit, version.Commit())
	}
	if body.Commit != "" && !regexp.MustCompile(`^[0-9a-f]{7,40}$`).MatchString(body.Commit) {
		t.Fatalf("commit = %q, want a short git hash", body.Commit)
	}
}

func TestLiveStartContract(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)
	streams.startResult = streammanager.Result{ID: "CMAG-000001-ABCDE", PlaylistURL: "/hls/CMAG-000001-ABCDE/index.m3u8"}

	// Act
	response := request(t, handler, http.MethodPost, "/api/live/start", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code","cameraIndex":1,"quality":4}`)

	// Assert
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	var body streammanager.Result
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ID != streams.startResult.ID || body.PlaylistURL != streams.startResult.PlaylistURL {
		t.Fatalf("body=%+v", body)
	}
}

func TestLiveStartDoesNotEchoSecret(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)
	secret := "do-not-echo-this"

	// Act
	response := request(t, handler, http.MethodPost, "/api/live/start", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"`+secret+`","cameraIndex":1,"quality":4}`)

	// Assert
	if strings.Contains(response.Body.String(), secret) {
		t.Fatal("response leaked security code")
	}
}

func TestLiveStartForwardsStreamParameters(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)

	// Act
	request(t, handler, http.MethodPost, "/api/live/start", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code","cameraIndex":1,"quality":4}`)

	// Assert
	if streams.startRequest.channel != 1 || streams.startRequest.quality != 4 {
		t.Fatalf("channel/quality = %d/%d", streams.startRequest.channel, streams.startRequest.quality)
	}
}

func TestLiveStartPassesSecurityCodeToSession(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)
	secret := "do-not-echo-this"

	// Act
	request(t, handler, http.MethodPost, "/api/live/start", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"`+secret+`","cameraIndex":1,"quality":4}`)

	// Assert
	if streams.startRequest.code != secret {
		t.Fatal("handler did not pass the security code to the session boundary")
	}
}

func TestLiveStartRejectsInputsOutsideSpecification(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)
	tests := []struct {
		name string
		body string
	}{
		{"empty", `{}`},
		{"invalid ip", `{"ip":"not-an-ip","did":"CMAG-000001-ABCDE","cameraIndex":0,"quality":1}`},
		{"empty did", `{"ip":"192.0.2.10","did":"","cameraIndex":0,"quality":1}`},
		{"camera below range", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","cameraIndex":-1,"quality":1}`},
		{"camera above range", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","cameraIndex":4,"quality":1}`},
		{"quality below range", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","cameraIndex":0,"quality":-1}`},
		{"quality above range", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","cameraIndex":0,"quality":5}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			response := request(t, handler, http.MethodPost, "/api/live/start", tt.body)

			// Assert
			if response.Code != http.StatusBadRequest {
				t.Errorf("status=%d, want 400", response.Code)
			}
		})
	}
}

func TestAuthenticationContract(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	s.authFn = func(context.Context, string, string, string) (ppcs.AuthResult, error) {
		return ppcs.AuthResult{DID: "CMAG-000001-ABCDE", Authenticated: true}, nil
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/authenticate", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code"}`)

	// Assert
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestAuthenticationDoesNotEchoSecret(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	secret := "test-secret"
	s.authFn = func(context.Context, string, string, string) (ppcs.AuthResult, error) {
		return ppcs.AuthResult{DID: "CMAG-000001-ABCDE", Authenticated: true}, nil
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/authenticate", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"`+secret+`"}`)

	// Assert
	if strings.Contains(response.Body.String(), secret) {
		t.Fatal("authentication response leaked secret")
	}
}

func TestAuthenticationPassesSecurityCodeToSession(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	secret := "test-secret"
	seen := ""
	s.authFn = func(_ context.Context, _, _, code string) (ppcs.AuthResult, error) {
		seen = code
		return ppcs.AuthResult{DID: "CMAG-000001-ABCDE", Authenticated: true}, nil
	}

	// Act
	request(t, handler, http.MethodPost, "/api/session/authenticate", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"`+secret+`"}`)

	// Assert
	if seen != secret {
		t.Fatalf("session received %q, want the request secret", seen)
	}
}

func TestAuthenticationFailureUsesUnauthorized(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	s.authFn = func(context.Context, string, string, string) (ppcs.AuthResult, error) {
		return ppcs.AuthResult{}, ppcs.ErrAuthRejected
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/authenticate", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"wrong"}`)

	// Assert
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", response.Code)
	}
}

func TestAuthenticationRetriesTransportErrors(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	var calls atomic.Int32
	s.authFn = func(context.Context, string, string, string) (ppcs.AuthResult, error) {
		if calls.Add(1) < 3 {
			return ppcs.AuthResult{}, ppcs.ErrLanTimeout
		}
		return ppcs.AuthResult{Authenticated: true}, nil
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/authenticate", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code"}`)

	// Assert
	if response.Code != http.StatusOK || calls.Load() != 3 {
		t.Fatalf("status=%d calls=%d", response.Code, calls.Load())
	}
}

func TestAuthenticationDoesNotRetryExplicitRejection(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	var calls atomic.Int32
	s.authFn = func(context.Context, string, string, string) (ppcs.AuthResult, error) {
		calls.Add(1)
		return ppcs.AuthResult{}, ppcs.ErrAuthRejected
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/authenticate", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"wrong"}`)

	// Assert
	if response.Code != http.StatusUnauthorized || calls.Load() != 1 {
		t.Fatalf("status=%d calls=%d", response.Code, calls.Load())
	}
}

func TestSessionControlRequestsAreSerialized(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	var active atomic.Int32
	var maximum atomic.Int32
	s.probeFn = func(context.Context, string, string) (ppcs.ProbeResult, error) {
		current := active.Add(1)
		for {
			old := maximum.Load()
			if current <= old || maximum.CompareAndSwap(old, current) {
				break
			}
		}
		time.Sleep(25 * time.Millisecond)
		active.Add(-1)
		return ppcs.ProbeResult{Ready: true}, nil
	}
	done := make(chan struct{}, 2)

	// Act
	for range 2 {
		go func() {
			request(t, handler, http.MethodPost, "/api/session/probe", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE"}`)
			done <- struct{}{}
		}()
	}
	<-done
	<-done

	// Assert
	if maximum.Load() != 1 {
		t.Fatalf("maximum concurrent session controls=%d, want 1", maximum.Load())
	}
}

func TestLiveStatusRequiresID(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)

	// Act
	response := request(t, handler, http.MethodGet, "/api/live/status", "")

	// Assert
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", response.Code)
	}
}

func TestLiveStatusReturnsBattery(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)
	streams.status = streammanager.Status{ID: "device", Running: true, Channel: 1, BitrateBytesPerSecond: 12345, Battery: &streammanager.Battery{Capacity: 77}}

	// Act
	response := request(t, handler, http.MethodGet, "/api/live/status?id=device", "")

	// Assert
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"capacity":77`) {
		t.Fatalf("unexpected status: %d %s", response.Code, response.Body.String())
	}
}

func TestLiveStopIsIdempotent(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)

	// Act
	first := request(t, handler, http.MethodPost, "/api/live/stop", `{"id":"device"}`)
	second := request(t, handler, http.MethodPost, "/api/live/stop", `{"id":"device"}`)

	// Assert
	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("status=%d/%d, want 200/200", first.Code, second.Code)
	}
	if len(streams.stopped) != 2 {
		t.Fatalf("stop calls=%d, want 2", len(streams.stopped))
	}
}

func TestMethodContract(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)

	// Act
	response := request(t, handler, http.MethodGet, "/api/live/start", "")

	// Assert
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d, want 405", response.Code)
	}
	if got := response.Header().Get("Allow"); got != "POST" {
		t.Fatalf("Allow=%q, want POST", got)
	}
}

func TestUnknownAPIIsJSON404NotStaticFileFallback(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)

	// Act
	response := request(t, handler, http.MethodGet, "/api/does-not-exist", "")

	// Assert
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404", response.Code)
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("content type=%q", got)
	}
}

func TestErrorResponsesCarryStableCodes(t *testing.T) {
	// Arrange
	_, handler, streams := testServer(t)
	streams.active = 1
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{"invalid json", http.MethodPost, "/api/channels", `{`, http.StatusBadRequest, "invalid_json"},
		{"validation", http.MethodPost, "/api/channels", `{}`, http.StatusBadRequest, "validation_failed"},
		{"conflict", http.MethodPost, "/api/channels", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"code"}`, http.StatusConflict, "conflict"},
		{"not found", http.MethodGet, "/api/does-not-exist", "", http.StatusNotFound, "not_found"},
		{"method not allowed", http.MethodGet, "/api/live/start", "", http.StatusMethodNotAllowed, "method_not_allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			response := request(t, handler, tt.method, tt.path, tt.body)

			// Assert
			if response.Code != tt.wantStatus {
				t.Fatalf("status=%d, want %d", response.Code, tt.wantStatus)
			}
			_, code := decodeError(t, response)
			if code != tt.wantCode {
				t.Fatalf("code=%q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestAuthenticationRejectionCarriesUnauthorizedCode(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	s.authFn = func(context.Context, string, string, string) (ppcs.AuthResult, error) {
		return ppcs.AuthResult{}, ppcs.ErrAuthRejected
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/authenticate", `{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"wrong"}`)

	// Assert
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", response.Code)
	}
	_, code := decodeError(t, response)
	if code != "unauthorized" {
		t.Fatalf("code=%q, want unauthorized", code)
	}
}

func TestProbeTimeoutCarriesUpstreamTimeoutCode(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	s.probeFn = func(context.Context, string, string) (ppcs.ProbeResult, error) {
		return ppcs.ProbeResult{}, ppcs.ErrLanTimeout
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/probe", `{"ip":"192.0.2.10"}`)

	// Assert
	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf("status=%d, want 504", response.Code)
	}
	_, code := decodeError(t, response)
	if code != "upstream_timeout" {
		t.Fatalf("code=%q, want upstream_timeout", code)
	}
}

func TestUnknownErrorUsesInternalError(t *testing.T) {
	// Arrange
	s, handler, _ := testServer(t)
	s.probeFn = func(context.Context, string, string) (ppcs.ProbeResult, error) {
		return ppcs.ProbeResult{}, errors.New("disk is on fire")
	}

	// Act
	response := request(t, handler, http.MethodPost, "/api/session/probe", `{"ip":"192.0.2.10"}`)

	// Assert
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500", response.Code)
	}
	message, code := decodeError(t, response)
	if code != "internal_error" {
		t.Fatalf("code=%q, want internal_error", code)
	}
	if strings.Contains(message, "disk is on fire") {
		t.Fatalf("message=%q leaks internal detail", message)
	}
}
