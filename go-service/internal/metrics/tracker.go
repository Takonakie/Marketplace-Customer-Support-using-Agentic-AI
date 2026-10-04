package metrics

import (
	"sync"
	"time"
)

type ToolCallLog struct {
	Tool       string  `json:"tool"`
	DurationMS float64 `json:"duration_ms"`
}

type ErrorLog struct {
	Error     string  `json:"error"`
	Timestamp float64 `json:"timestamp"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type TraceLog struct {
	TraceID    string        `json:"trace_id"`
	MessageID  string        `json:"message_id"`
	ChatID     int64         `json:"chat_id"`
	LatencyMS  float64       `json:"latency_ms"`
	TokenUsage TokenUsage    `json:"token_usage"`
	ToolCalls  []ToolCallLog `json:"tool_calls"`
	Errors     []ErrorLog    `json:"errors"`
	Timestamp  time.Time     `json:"timestamp"`
}

type SystemMetrics struct {
	TotalRequests     int64            `json:"total_requests"`
	TotalTokens       int64            `json:"total_tokens"`
	TotalPromptTokens int64            `json:"total_prompt_tokens"`
	TotalComplTokens  int64            `json:"total_completion_tokens"`
	AvgLatencyMS      float64          `json:"avg_latency_ms"`
	TotalErrors       int64            `json:"total_errors"`
	ToolCallCounts    map[string]int64 `json:"tool_call_counts"`
	RecentTraces      []TraceLog       `json:"recent_traces"`
}

type Tracker struct {
	mu           sync.RWMutex
	totalReqs    int64
	totalTokens  int64
	totalPrompt  int64
	totalCompl   int64
	sumLatency   float64
	totalErrors  int64
	toolCounts   map[string]int64
	recentTraces []TraceLog
}

var GlobalTracker = NewTracker()

func NewTracker() *Tracker {
	return &Tracker{
		toolCounts:   make(map[string]int64),
		recentTraces: make([]TraceLog, 0),
	}
}

func (t *Tracker) RecordTrace(trace TraceLog) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.totalReqs++
	t.sumLatency += trace.LatencyMS

	tot := trace.TokenUsage.PromptTokens + trace.TokenUsage.CompletionTokens
	t.totalTokens += int64(tot)
	t.totalPrompt += int64(trace.TokenUsage.PromptTokens)
	t.totalCompl += int64(trace.TokenUsage.CompletionTokens)

	if len(trace.Errors) > 0 {
		t.totalErrors += int64(len(trace.Errors))
	}

	for _, tc := range trace.ToolCalls {
		t.toolCounts[tc.Tool]++
	}

	if trace.Timestamp.IsZero() {
		trace.Timestamp = time.Now()
	}

	t.recentTraces = append([]TraceLog{trace}, t.recentTraces...)
	if len(t.recentTraces) > 100 {
		t.recentTraces = t.recentTraces[:100]
	}
}

func (t *Tracker) GetMetrics() SystemMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	avgLat := 0.0
	if t.totalReqs > 0 {
		avgLat = t.sumLatency / float64(t.totalReqs)
	}

	toolCountsCopy := make(map[string]int64)
	for k, v := range t.toolCounts {
		toolCountsCopy[k] = v
	}

	tracesCopy := make([]TraceLog, len(t.recentTraces))
	copy(tracesCopy, t.recentTraces)

	return SystemMetrics{
		TotalRequests:     t.totalReqs,
		TotalTokens:       t.totalTokens,
		TotalPromptTokens: t.totalPrompt,
		TotalComplTokens:  t.totalCompl,
		AvgLatencyMS:      avgLat,
		TotalErrors:       t.totalErrors,
		ToolCallCounts:    toolCountsCopy,
		RecentTraces:      tracesCopy,
	}
}
