package bot

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/agenticsdk/go-service/internal/metrics"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

type TelegramMessage struct {
	MessageID    string `json:"message_id"`
	ChatID       int64  `json:"chat_id"`
	CustomerName string `json:"customer_name"`
	Text         string `json:"text"`
	Timestamp    string `json:"timestamp"`
}

type TelegramResponse struct {
	MessageID string                 `json:"message_id"`
	ChatID    int64                  `json:"chat_id"`
	ReplyText string                 `json:"reply_text"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type BotService struct {
	bot *tgbotapi.BotAPI
	rdb *redis.Client
}

func NewBotService(token string, rdb *redis.Client) (*BotService, error) {
	if token == "" {
		log.Println("TELEGRAM_BOT_TOKEN is empty; bot will operate in mock mode for Redis pub/sub")
		return &BotService{bot: nil, rdb: rdb}, nil
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &BotService{bot: bot, rdb: rdb}, nil
}

func (s *BotService) Start(ctx context.Context) {
	go s.listenOutgoing(ctx)

	if s.bot == nil {
		log.Println("Telegram Bot API disabled (no token). Listening only to Redis outgoing_messages.")
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := s.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := TelegramMessage{
			MessageID:    strconv.Itoa(update.Message.MessageID),
			ChatID:       update.Message.Chat.ID,
			CustomerName: update.Message.From.FirstName,
			Text:         update.Message.Text,
			Timestamp:    update.Message.Time().Format("2006-01-02T15:04:05Z07:00"),
		}

		data, _ := json.Marshal(msg)
		s.rdb.Publish(ctx, "incoming_messages", data)
		log.Printf("Published message from chat %d to Redis", msg.ChatID)
	}
}

func (s *BotService) listenOutgoing(ctx context.Context) {
	sub := s.rdb.Subscribe(ctx, "outgoing_messages")
	ch := sub.Channel()

	for msg := range ch {
		var resp TelegramResponse
		if err := json.Unmarshal([]byte(msg.Payload), &resp); err != nil {
			log.Printf("Error unmarshalling outgoing msg: %v", err)
			continue
		}

		// Record monitoring trace metrics
		trace := parseTraceLog(resp)
		metrics.GlobalTracker.RecordTrace(trace)

		log.Printf("Received response for chat %d: %s", resp.ChatID, resp.ReplyText)
		if s.bot != nil {
			tgMsg := tgbotapi.NewMessage(resp.ChatID, resp.ReplyText)
			tgMsg.ParseMode = "Markdown"
			if _, err := s.bot.Send(tgMsg); err != nil {
				log.Printf("[WARN] Failed to send Markdown message to chat %d (%v), retrying without ParseMode...", resp.ChatID, err)
				tgMsg.ParseMode = ""
				if _, err := s.bot.Send(tgMsg); err != nil {
					log.Printf("[ERROR] Failed to send Telegram message to chat %d: %v", resp.ChatID, err)
				} else {
					log.Printf("Successfully sent plain text Telegram message to chat %d", resp.ChatID)
				}
			} else {
				log.Printf("Successfully sent Markdown Telegram message to chat %d", resp.ChatID)
			}
		}
	}
}

func parseTraceLog(resp TelegramResponse) metrics.TraceLog {
	trace := metrics.TraceLog{
		MessageID:  resp.MessageID,
		ChatID:     resp.ChatID,
		AIResponse: resp.ReplyText,
		Timestamp:  time.Now(),
	}

	meta := resp.Metadata
	if meta == nil {
		return trace
	}

	if tid, ok := meta["trace_id"].(string); ok {
		trace.TraceID = tid
	}
	if lat, ok := meta["latency_ms"].(float64); ok {
		trace.LatencyMS = lat
	}
	if um, ok := meta["user_message"].(string); ok {
		trace.UserMessage = um
	}
	if air, ok := meta["ai_response"].(string); ok && air != "" {
		trace.AIResponse = air
	}

	if tu, ok := meta["token_usage"].(map[string]interface{}); ok {
		if pt, ok := tu["prompt_tokens"].(float64); ok {
			trace.TokenUsage.PromptTokens = int(pt)
		}
		if ct, ok := tu["completion_tokens"].(float64); ok {
			trace.TokenUsage.CompletionTokens = int(ct)
		}
	}

	if tcs, ok := meta["tool_calls"].([]interface{}); ok {
		for _, rawTc := range tcs {
			if tcMap, ok := rawTc.(map[string]interface{}); ok {
				var tc metrics.ToolCallLog
				if tName, ok := tcMap["tool"].(string); ok {
					tc.Tool = tName
				}
				if dur, ok := tcMap["duration_ms"].(float64); ok {
					tc.DurationMS = dur
				}
				trace.ToolCalls = append(trace.ToolCalls, tc)
			}
		}
	}

	if errs, ok := meta["errors"].([]interface{}); ok {
		for _, rawErr := range errs {
			if errMap, ok := rawErr.(map[string]interface{}); ok {
				var el metrics.ErrorLog
				if eMsg, ok := errMap["error"].(string); ok {
					el.Error = eMsg
				}
				if ts, ok := errMap["timestamp"].(float64); ok {
					el.Timestamp = ts
				}
				trace.Errors = append(trace.Errors, el)
			}
		}
	}

	return trace
}
