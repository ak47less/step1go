package restful

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ak47less/step1go/lite2/core"
	"github.com/ak47less/step1go/lite2/lib"
)

type fakeActionController struct {
	context  *lib.ActionContext
	calls    []string
	seekTo   *lib.RelativePosition
	playErr  error
	startErr error
}

func (inst *fakeActionController) NewContext() *lib.ActionContext {
	return inst.context
}

func (inst *fakeActionController) StartEngine(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "start_engine")
	return inst.startErr
}

func (inst *fakeActionController) StopEngine(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "stop_engine")
	return nil
}

func (inst *fakeActionController) RestartEngine(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "restart_engine")
	return nil
}

func (inst *fakeActionController) Play(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "play")
	return inst.playErr
}

func (inst *fakeActionController) Stop(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "stop")
	return nil
}

func (inst *fakeActionController) Pause(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "pause")
	return nil
}

func (inst *fakeActionController) Resume(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "resume")
	return nil
}

func (inst *fakeActionController) Record(*lib.ActionContext) error {
	inst.calls = append(inst.calls, "record")
	return nil
}

func (inst *fakeActionController) Seek(_ *lib.ActionContext, position *lib.RelativePosition) error {
	inst.calls = append(inst.calls, "seek")
	inst.seekTo = position.Clone()
	return nil
}

type fakeTransport struct {
	position *lib.RelativePosition
}

func (inst *fakeTransport) Play() error {
	return nil
}

func (inst *fakeTransport) Record() error {
	return nil
}

func (inst *fakeTransport) Stop() error {
	return nil
}

func (inst *fakeTransport) Pause() error {
	return nil
}

func (inst *fakeTransport) Resume() error {
	return nil
}

func (inst *fakeTransport) Seek(*lib.RelativePosition) error {
	return nil
}

func (inst *fakeTransport) GetCurrentPosition() *lib.RelativePosition {
	return inst.position
}

func newTestWebServer(t *testing.T) (*innerWebServer, *fakeActionController) {
	t.Helper()

	document := &core.Timeline{
		Name:   "@document",
		Type:   core.TimelineTypeDocument,
		Length: 60,
	}
	engineContext := &lib.EngineContext{
		Started:           true,
		DocumentTimeline:  document,
		EngineTimeline:    &core.Timeline{Name: "@engine", Type: core.TimelineTypeEngine},
		TransportTimeline: &core.Timeline{Name: "@transport", Type: core.TimelineTypeTransport},
		StepTimeline:      &core.Timeline{Name: "@step", Type: core.TimelineTypeStep},
		FrameTimeline:     &core.Timeline{Name: "@frame", Type: core.TimelineTypeFrame},
	}
	transportContext := &lib.TransportContext{
		Playing: true,
	}
	transport := &fakeTransport{
		position: &lib.RelativePosition{
			Ref: document,
			Pos: 12.5,
		},
	}

	cube := new(lib.Cube)
	err := cube.Init(&lib.CubeConfig{
		BankListCapacity:  2,
		BankListLength:    2,
		StepListCapacity:  2,
		StepListLength:    2,
		TrackListCapacity: 2,
		TrackListLength:   2,
	})
	if err != nil {
		t.Fatalf("init cube: %v", err)
	}

	context := &lib.ApplicationContext{
		EngineContext:    engineContext,
		TransportContext: transportContext,
		SequencerContext: &lib.SequencerContext{Cube: *cube},
		Transport:        transport,
	}
	transportContext.AC = context
	transportContext.Transport = transport
	engineContext.AC = context

	controller := &fakeActionController{
		context: &lib.ActionContext{
			AC: context,
			EC: engineContext,
		},
	}
	context.ActionController = controller

	web := new(innerWebServer)
	if err := web.init(context, &lib.WebServerConfig{}); err != nil {
		t.Fatalf("init web server: %v", err)
	}
	return web, controller
}

func performRequest(handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, target, reader)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeErrorCode(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var payload apiErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return payload.Error.Code
}

func TestHealthAndErrorResponses(t *testing.T) {
	web, _ := newTestWebServer(t)
	handler := web.handler()

	response := performRequest(handler, http.MethodGet, "/api/v1/health", "")
	if response.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", response.Code, http.StatusOK)
	}
	var health struct {
		OK      bool   `json:"ok"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if !health.OK || health.Version == "" {
		t.Fatalf("unexpected health response: %+v", health)
	}

	response = performRequest(handler, http.MethodPost, "/api/v1/health", "{}")
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if code := decodeErrorCode(t, response); code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("error code = %q", code)
	}

	response = performRequest(handler, http.MethodGet, "/api/v1/missing", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("not-found status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if code := decodeErrorCode(t, response); code != "NOT_FOUND" {
		t.Fatalf("error code = %q", code)
	}
}

func TestSPAIsMountedAtRoot(t *testing.T) {
	web, _ := newTestWebServer(t)
	handler := web.handler()

	for _, target := range []string{"/", "/projects/current"} {
		response := performRequest(handler, http.MethodGet, target, "")
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", target, response.Code, http.StatusOK)
		}
		if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
			t.Fatalf("%s content type = %q", target, contentType)
		}
		if !strings.Contains(response.Body.String(), "<!DOCTYPE html>") {
			t.Fatalf("%s did not return index.html", target)
		}
	}

	response := performRequest(handler, http.MethodGet, "/missing.js", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing asset status = %d, want %d", response.Code, http.StatusNotFound)
	}

	response = performRequest(handler, http.MethodPost, "/projects/current", "{}")
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("SPA POST status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func TestStatusAndCommands(t *testing.T) {
	web, controller := newTestWebServer(t)
	handler := web.handler()

	response := performRequest(handler, http.MethodGet,
		"/api/v1/status?keys=engine.started,transport.playing,unknown.key", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d", response.Code)
	}
	var statusResponse struct {
		Status map[string]any `json:"status"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &statusResponse); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if statusResponse.Status["engine.started"] != float64(1) ||
		statusResponse.Status["transport.playing"] != float64(1) {
		t.Fatalf("unexpected status: %#v", statusResponse.Status)
	}
	if _, ok := statusResponse.Status["unknown.key"]; ok {
		t.Fatal("unknown status key must be omitted")
	}

	response = performRequest(handler, http.MethodPost, "/api/v1/engine/start", "{}")
	if response.Code != http.StatusOK {
		t.Fatalf("start engine code = %d", response.Code)
	}
	if len(controller.calls) != 1 || controller.calls[0] != "start_engine" {
		t.Fatalf("controller calls = %v", controller.calls)
	}

	response = performRequest(handler, http.MethodPut, "/api/v1/transport/position",
		`{"position":2.5,"timeline":"@document"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("seek code = %d, body = %s", response.Code, response.Body.String())
	}
	if controller.seekTo == nil || float64(controller.seekTo.Pos) != 2.5 ||
		controller.seekTo.Ref != controller.context.EC.DocumentTimeline {
		t.Fatalf("unexpected seek target: %+v", controller.seekTo)
	}
}

func TestCubeResourcesAndPatch(t *testing.T) {
	web, _ := newTestWebServer(t)
	handler := web.handler()

	response := performRequest(handler, http.MethodGet, "/api/v1/cube", "")
	if response.Code != http.StatusOK {
		t.Fatalf("cube code = %d", response.Code)
	}
	var cube cubeResource
	if err := json.Unmarshal(response.Body.Bytes(), &cube); err != nil {
		t.Fatalf("decode cube: %v", err)
	}
	if cube.Banks.Length != 2 || cube.Steps.Length != 2 || cube.Tracks.Length != 2 {
		t.Fatalf("unexpected cube: %+v", cube)
	}

	response = performRequest(handler, http.MethodPatch, "/api/v1/tracks/0",
		`{"label":"Bass","enabled":true,"gainDb":-6,"channel":-1}`)
	if response.Code != http.StatusOK {
		t.Fatalf("patch track code = %d, body = %s", response.Code, response.Body.String())
	}
	var track trackResource
	if err := json.Unmarshal(response.Body.Bytes(), &track); err != nil {
		t.Fatalf("decode track: %v", err)
	}
	if track.Label != "Bass" || !track.Enabled || track.GainDB != -6 || track.Channel != -1 {
		t.Fatalf("unexpected track: %+v", track)
	}

	response = performRequest(handler, http.MethodPatch, "/api/v1/banks/0",
		`{"label":"Verse"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("patch bank code = %d", response.Code)
	}

	response = performRequest(handler, http.MethodGet,
		"/api/v1/cells?bankIndex=1&stepIndex=0&trackIndex=1", "")
	if response.Code != http.StatusOK {
		t.Fatalf("cells code = %d", response.Code)
	}
	var cells cellListResource
	if err := json.Unmarshal(response.Body.Bytes(), &cells); err != nil {
		t.Fatalf("decode cells: %v", err)
	}
	if len(cells.Items) != 1 || cells.Items[0].Name != "b1_s0_t1" {
		t.Fatalf("unexpected cells: %+v", cells.Items)
	}

	response = performRequest(handler, http.MethodGet, "/api/v1/cells/b1_s0_t1", "")
	if response.Code != http.StatusOK {
		t.Fatalf("cell code = %d", response.Code)
	}
}

func TestBatchActions(t *testing.T) {
	web, controller := newTestWebServer(t)
	handler := web.handler()

	body := `{
		"actions": [
			{"name": "start_engine", "params": {}},
			{"name": "status", "all": false, "keys": ["engine.started"]}
		]
	}`
	response := performRequest(handler, http.MethodPost, "/api/v1/actions", body)
	if response.Code != http.StatusOK {
		t.Fatalf("actions code = %d, body = %s", response.Code, response.Body.String())
	}
	var result batchResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode actions: %v", err)
	}
	if len(result.Results) != 2 || !result.Results[0].OK || !result.Results[1].OK {
		t.Fatalf("unexpected action results: %+v", result.Results)
	}
	if result.Status["engine.started"] != float64(1) {
		t.Fatalf("unexpected action status: %#v", result.Status)
	}
	if len(controller.calls) != 1 || controller.calls[0] != "start_engine" {
		t.Fatalf("controller calls = %v", controller.calls)
	}

	controller.playErr = errors.New("engine must not be started while playing")
	response = performRequest(handler, http.MethodPost, "/api/v1/actions",
		`{"actions":[{"name":"play"}]}`)
	if response.Code != http.StatusConflict {
		t.Fatalf("failed action code = %d, want %d", response.Code, http.StatusConflict)
	}
	if code := decodeErrorCode(t, response); code != "INVALID_STATE" {
		t.Fatalf("failed action code = %q", code)
	}

	response = performRequest(handler, http.MethodPost, "/api/v1/actions",
		`{"actions":[{"name":"unknown"}]}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unknown action code = %d", response.Code)
	}
	if code := decodeErrorCode(t, response); code != "UNKNOWN_ACTION" {
		t.Fatalf("unknown action code = %q", code)
	}
}

func TestServerStartAndStop(t *testing.T) {
	web, _ := newTestWebServer(t)
	web.host = "127.0.0.1"
	web.port = 0

	if err := web.start(); err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("network sockets are unavailable in this environment: %v", err)
		}
		t.Fatalf("start server: %v", err)
	}
	address := web.serverAddress()
	if address == "" {
		t.Fatal("server address is empty")
	}

	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://" + address + "/api/v1/health")
	if err != nil {
		t.Fatalf("request health: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", response.StatusCode)
	}
	response.Body.Close()

	response, err = client.Get("http://" + address + "/")
	if err != nil {
		t.Fatalf("request SPA: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("SPA status = %d", response.StatusCode)
	}
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("SPA content type = %q", contentType)
	}

	if err := web.stop(); err != nil {
		t.Fatalf("stop server: %v", err)
	}
	if err := web.stop(); err != nil {
		t.Fatalf("stop server twice: %v", err)
	}
}
