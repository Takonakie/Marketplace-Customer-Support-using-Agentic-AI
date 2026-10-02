import json
import asyncio
import redis
from app.config import settings
from app.chat_history.manager import ChatHistoryManager
from app.agent.core import handle_message

def start_worker():
    rdb = redis.Redis.from_url(settings.REDIS_URL, decode_responses=True)
    history_manager = ChatHistoryManager(rdb)
    pubsub = rdb.pubsub()
    pubsub.subscribe("incoming_messages")

    print(f"Subscribed to 'incoming_messages' on Redis ({settings.REDIS_URL})...")
    for item in pubsub.listen():
        if item["type"] == "message":
            try:
                data = json.loads(item["data"])
                print(f"Received message from Redis: {data.get('text')}")
                
                # Run async handle_message
                loop = asyncio.get_event_loop()
                if loop.is_closed():
                    loop = asyncio.new_event_loop()
                    asyncio.set_event_loop(loop)

                resp = loop.run_until_complete(handle_message(data, history_manager))
                rdb.publish("outgoing_messages", json.dumps(resp))
                print(f"Published reply to 'outgoing_messages' for chat {data.get('chat_id')}")
            except Exception as e:
                print(f"Error processing Redis message: {e}")
