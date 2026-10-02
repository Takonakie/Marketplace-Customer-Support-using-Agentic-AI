import json
import redis
from app.config import settings

HISTORY_TTL = 86400          # 24 hours
MAX_HISTORY_LENGTH = 20      # 10 turns (user + assistant)

class ChatHistoryManager:
    def __init__(self, redis_client: redis.Redis):
        self.redis = redis_client

    def _key(self, chat_id: int) -> str:
        return f"chat_history:{chat_id}"

    def get_history(self, chat_id: int) -> list[dict]:
        key = self._key(chat_id)
        raw_messages = self.redis.lrange(key, -MAX_HISTORY_LENGTH, -1)
        return [json.loads(m) for m in raw_messages]

    def append_message(self, chat_id: int, role: str, content: str):
        key = self._key(chat_id)
        entry = json.dumps({"role": role, "content": content})
        self.redis.rpush(key, entry)
        self.redis.ltrim(key, -MAX_HISTORY_LENGTH, -1)
        self.redis.expire(key, HISTORY_TTL)

    def clear_history(self, chat_id: int):
        self.redis.delete(self._key(chat_id))
