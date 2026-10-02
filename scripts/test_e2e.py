import json
import time
import uuid
import redis
import os
from dotenv import load_dotenv

load_dotenv()
REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6380")

def send_and_receive(rdb: redis.Redis, pubsub, text: str, chat_id: int = 12345, customer_name: str = "Budi", timeout: int = 45):
    msg_id = str(uuid.uuid4())
    test_payload = {
        "message_id": msg_id,
        "chat_id": chat_id,
        "customer_name": customer_name,
        "text": text,
        "timestamp": "2026-10-02T22:00:00Z"
    }

    print(f"\n==================================================")
    print(f"Customer ({customer_name}): \"{text}\"")
    print(f"==================================================")
    
    # 1. Pastikan mengosongkan antrean lama sebelum mengirim pesan baru
    while pubsub.get_message(timeout=0.1):
        pass

    # 2. Publish pesan pengujian
    rdb.publish("incoming_messages", json.dumps(test_payload))
    start = time.time()

    # 3. Baca pesan balasan dari outgoing_messages
    while True:
        message = pubsub.get_message(timeout=1.0)
        if message and message["type"] == "message":
            # Pastikan channel balasan berasal dari outgoing_messages
            if message.get("channel") == "outgoing_messages":
                parsed = json.loads(message["data"])
                # Pastikan message_id sesuai dengan pesan yang baru dikirim
                if parsed.get("message_id") == msg_id or parsed.get("chat_id") == chat_id:
                    print(f"\n[Agent Response]:\n{parsed.get('reply_text')}\n")
                    metadata = parsed.get("metadata", {})
                    print(f"[Metadata]: Latency={metadata.get('latency_ms', 0):.2f}ms | Tools={metadata.get('tool_calls')} | Tokens={metadata.get('token_usage')}")
                    return parsed

        if time.time() - start > timeout:
            print(f"\n[ERROR] Timeout waiting for response to: '{text}' (after {timeout}s)")
            return None

def run_all_tests():
    rdb = redis.Redis.from_url(REDIS_URL, decode_responses=True)
    pubsub = rdb.pubsub()
    pubsub.subscribe("outgoing_messages")
    time.sleep(1) # Pastikan subscription aktif di Redis

    print(">>> Starting End-to-End Comprehensive Feature Suite <<<")

    # Menggunakan chat_id=12345 sesuai data seed (tg_12345)
    print("\n--- SKENARIO 1: Inquiry Total Pembelian ---")
    send_and_receive(rdb, pubsub, "Berapa total akumulasi pembelian saya tahun ini?", chat_id=12345, customer_name="Budi")

    print("\n--- SKENARIO 2: Inquiry Status Pengiriman ---")
    send_and_receive(rdb, pubsub, "Kok barang saya belum sampe ya?", chat_id=12345, customer_name="Budi")

    print("\n--- SKENARIO 3: Eskalasi Refund (Case Creation) ---")
    send_and_receive(rdb, pubsub, "Saya abis melakukan pembatalan order, kok duid saya belum refund ya?", chat_id=12345, customer_name="Budi")

    print("\n--- SKENARIO 4: Multi-Turn Conversation ---")
    send_and_receive(rdb, pubsub, "Bisa sebutkan daftar pesanan saya?", chat_id=12345, customer_name="Budi")
    time.sleep(1)
    send_and_receive(rdb, pubsub, "Tolong cek status yang Sepatu Badminton Li-Ning saja", chat_id=12345, customer_name="Budi")

    print("\n>>> All End-to-End Scenarios Completed! <<<")

if __name__ == "__main__":
    run_all_tests()
