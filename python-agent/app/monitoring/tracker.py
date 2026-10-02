import time
import uuid
import logging

logger = logging.getLogger("agent_tracker")

class MonitoringTracker:
    def __init__(self, message_id: str):
        self.message_id = message_id
        self.trace_id = str(uuid.uuid4())
        self.start_time = time.time()
        self.token_usage = {"prompt_tokens": 0, "completion_tokens": 0}
        self.tool_calls = []
        self.errors = []

    def record_llm_call(self, usage: dict):
        if usage:
            self.token_usage["prompt_tokens"] += getattr(usage, "prompt_tokens", 0)
            self.token_usage["completion_tokens"] += getattr(usage, "completion_tokens", 0)

    def record_tool_call(self, tool_name: str, duration_ms: float):
        self.tool_calls.append({"tool": tool_name, "duration_ms": duration_ms})

    def record_error(self, error: str):
        self.errors.append({"error": str(error), "timestamp": time.time()})

    def finalize(self) -> dict:
        latency_ms = (time.time() - self.start_time) * 1000
        metrics = {
            "message_id": self.message_id,
            "trace_id": self.trace_id,
            "latency_ms": latency_ms,
            "token_usage": self.token_usage,
            "tool_calls": self.tool_calls,
            "errors": self.errors
        }
        logger.info(f"MONITORING METRICS: {metrics}")
        return metrics
