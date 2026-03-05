package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	OllamaHost   string   `json:"ollama_host"`
	ServerHost   string   `json:"server_host"`
	ServerPort   string   `json:"server_port"`
	Models       []string `json:"preferred_models"`
	ModelTimeout int      `json:"model_timeout_seconds"`
}

type ModelStatus struct {
	Model      string    `json:"model"`
	Loaded     bool      `json:"loaded"`
	LoadedAt   time.Time `json:"loaded_at,omitempty"`
	LastUsedAt time.Time `json:"last_used_at"`
}

type ChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Stream bool `json:"stream,omitempty"`
}

type ChatResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

type IntentRequest struct {
	Query string `json:"query"`
}

type IntentResponse struct {
	Intent  string            `json:"intent"`
	Module  string            `json:"module"`
	Context map[string]string `json:"context"`
	Action  string            `json:"action"`
}

var featureMap = map[string][]string{
	"recon":    {"port_scanner", "os_fingerprint", "subdomain_enum", "dir_brute", "vuln_scan", "tech_detect", "ssl_check", "dns_enum", "whois_lookup", "snmp_walk"},
	"exploitd": {"payload_generator", "reverse_shell", "sql_injector", "xss_tester", "lfi_scanner", "rce_exploit", "priv_escalate", "cred_dumper", "hash_cracker", "brute_forcer"},
	"stealth":  {"log_wiper", "mac_changer", "proxy_chain", "vpn_manager", "tor_router", "timestomp", "process_hider", "file_obfuscator", "traffic_shaper", "anti_forensics"},
}

var (
	config          Config
	loadedModels    atomic.Pointer[map[string]ModelStatus]
	modelLock       sync.Mutex
	defaultModel    string
)

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	if config.OllamaHost == "" {
		config.OllamaHost = "localhost:11434"
	}
	if config.ServerHost == "" {
		config.ServerHost = "localhost"
	}
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}
	if config.ModelTimeout == 0 {
		config.ModelTimeout = 300
	}
	if len(config.Models) > 0 {
		defaultModel = config.Models[0]
	}
	emptyMap := make(map[string]ModelStatus)
	loadedModels.Store(&emptyMap)
	return nil
}

func getOllamaURL(path string) string {
	return fmt.Sprintf("http://%s%s", config.OllamaHost, path)
}

func checkModelLoaded(model string) bool {
	models := loadedModels.Load()
	status, ok := (*models)[model]
	return ok && status.Loaded
}

func setModelLoaded(model string, loaded bool) {
	modelLock.Lock()
	defer modelLock.Unlock()
	models := loadedModels.Load()
	if *models == nil {
		*models = make(map[string]ModelStatus)
	}
	newModels := *models
	if loaded {
		newModels[model] = ModelStatus{
			Model:      model,
			Loaded:     true,
			LoadedAt:   time.Now(),
			LastUsedAt: time.Now(),
		}
	} else {
		delete(newModels, model)
	}
	loadedModels.Store(&newModels)
}

func updateModelLastUsed(model string) {
	modelLock.Lock()
	defer modelLock.Unlock()
	models := loadedModels.Load()
	if status, ok := (*models)[model]; ok {
		status.LastUsedAt = time.Now()
		(*models)[model] = status
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getOllamaURL("/api/tags"), nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":        "unhealthy",
			"ollama_error":  err.Error(),
			"loaded_models": getLoadedModelsList(),
		})
		return
	}
	defer resp.Body.Close()

	models := loadedModels.Load()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "healthy",
		"ollama":        "connected",
		"loaded_models": *models,
	})
}

func loadModelHandler(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	if model == "" {
		model = defaultModel
	}
	if model == "" {
		http.Error(w, "model parameter required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	generateReq := map[string]interface{}{
		"model":  model,
		"prompt": "test",
		"stream": false,
	}

	jsonData, err := json.Marshal(generateReq)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, getOllamaURL("/api/generate"), io.NopCloser(bytes.NewReader(jsonData)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load model: %v", err), http.StatusGatewayUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Ollama error: %s", body), http.StatusBadGateway)
		return
	}

	setModelLoaded(model, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "loaded",
		"model":  model,
	})
}

func unloadModelHandler(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	if model == "" {
		http.Error(w, "model parameter required", http.StatusBadRequest)
		return
	}

	setModelLoaded(model, false)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "unloaded",
		"model":  model,
	})
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	var chatReq ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if chatReq.Model == "" {
		chatReq.Model = defaultModel
	}
	if chatReq.Model == "" {
		http.Error(w, "no model specified", http.StatusBadRequest)
		return
	}

	if !checkModelLoaded(chatReq.Model) {
		http.Error(w, "model not loaded", http.StatusPreconditionFailed)
		return
	}

	updateModelLastUsed(chatReq.Model)

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
	defer cancel()

	ollamaReq := map[string]interface{}{
		"model":    chatReq.Model,
		"messages": chatReq.Messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, getOllamaURL("/api/chat"), io.NopCloser(bytes.NewReader(jsonData)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ollama error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Ollama error: %s", body), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func chatStreamHandler(w http.ResponseWriter, r *http.Request) {
	var chatReq ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if chatReq.Model == "" {
		chatReq.Model = defaultModel
	}
	if chatReq.Model == "" {
		http.Error(w, "no model specified", http.StatusBadRequest)
		return
	}

	if !checkModelLoaded(chatReq.Model) {
		http.Error(w, "model not loaded", http.StatusPreconditionFailed)
		return
	}

	updateModelLastUsed(chatReq.Model)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ollamaReq := map[string]interface{}{
		"model":    chatReq.Model,
		"messages": chatReq.Messages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, getOllamaURL("/api/chat"), io.NopCloser(bytes.NewReader(jsonData)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ollama error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	io.Copy(w, resp.Body)
}

func getLoadedModelsList() []string {
	models := loadedModels.Load()
	list := make([]string, 0, len(*models))
	for model := range *models {
		list = append(list, model)
	}
	return list
}

func gatherAutoContext() map[string]string {
	cwd, _ := os.Getwd()
	ctx := map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cwd":  cwd,
		"time": time.Now().Format(time.RFC3339),
	}
	if user := os.Getenv("USER"); user != "" {
		ctx["user"] = user
	} else if user := os.Getenv("USERNAME"); user != "" {
		ctx["user"] = user
	}
	return ctx
}

func intentParserHandler(w http.ResponseWriter, r *http.Request) {
	var req IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctxData := gatherAutoContext()
	ctxBytes, _ := json.Marshal(ctxData)
	ctxStr := string(ctxBytes)

	queryLower := strings.ToLower(req.Query)
	intent := "unknown"
	module := "unknown"

	// Routing logic for the 30 new features
	for category, features := range featureMap {
		if strings.Contains(queryLower, category) {
			intent = category
		}
		for _, feature := range features {
			if strings.Contains(queryLower, strings.ReplaceAll(feature, "_", " ")) || strings.Contains(queryLower, feature) {
				intent = category
				module = feature
				break
			}
		}
		if intent != "unknown" && module != "unknown" {
			break
		}
	}

	if intent == "unknown" {
		if strings.Contains(queryLower, "scan") || strings.Contains(queryLower, "find") || strings.Contains(queryLower, "discover") {
			intent = "recon"
			module = "port_scanner"
		} else if strings.Contains(queryLower, "hack") || strings.Contains(queryLower, "attack") || strings.Contains(queryLower, "exploit") {
			intent = "exploitd"
			module = "payload_generator"
		} else if strings.Contains(queryLower, "hide") || strings.Contains(queryLower, "cover") || strings.Contains(queryLower, "stealth") {
			intent = "stealth"
			module = "log_wiper"
		}
	}

	resp := IntentResponse{
		Intent:  intent,
		Module:  module,
		Context: ctxData,
		Action:  fmt.Sprintf("Routed to %s daemon with context size %d", intent, len(ctxStr)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := loadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler)
	r.Post("/load-model", loadModelHandler)
	r.Post("/unload-model", unloadModelHandler)
	r.Post("/chat", chatHandler)
	r.Post("/chat/stream", chatStreamHandler)
	r.Post("/intent_parser", intentParserHandler)

	addr := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting ai-core daemon on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
