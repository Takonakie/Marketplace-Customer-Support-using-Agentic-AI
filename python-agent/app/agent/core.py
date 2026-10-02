import json
import time
from openai import OpenAI
from app.config import settings
from app.agent.prompts import build_system_prompt
from app.agent.tools import TOOL_DEFINITIONS, TOOL_MAP
from app.agent.guardrails import validate_input, sanitize_output
from app.monitoring.tracker import MonitoringTracker
from app.chat_history.manager import ChatHistoryManager

async def handle_message(message: dict, history_manager: ChatHistoryManager) -> dict:
    tracker = MonitoringTracker(message.get("message_id", ""))
    chat_id = message["chat_id"]
    customer_name = message.get("customer_name", "Customer")
    text = message.get("text", "")

    guard_err = validate_input(text)
    if guard_err:
        return {
            "message_id": message.get("message_id"),
            "chat_id": chat_id,
            "reply_text": guard_err,
            "metadata": tracker.finalize()
        }

    system_prompt = build_system_prompt(chat_id, customer_name)
    chat_history = history_manager.get_history(chat_id)

    messages = [
        {"role": "system", "content": system_prompt},
        *chat_history,
        {"role": "user", "content": text}
    ]

    if not settings.LLM_API_KEY:
        # Development mode fallback when API Key is absent
        reply = f"Hai {customer_name}, sistem menerima pesan Anda: '{text}'. (Dev Mode - API Key belum diisi)"
        history_manager.append_message(chat_id, "user", text)
        history_manager.append_message(chat_id, "assistant", reply)
        return {
            "message_id": message.get("message_id"),
            "chat_id": chat_id,
            "reply_text": reply,
            "metadata": tracker.finalize()
        }

    # Inisialisasi OpenAI client dengan custom base_url jika ada
    client_kwargs = {"api_key": settings.LLM_API_KEY}
    if settings.LLM_BASE_URL:
        client_kwargs["base_url"] = settings.LLM_BASE_URL

    client = OpenAI(**client_kwargs)

    try:
        while True:
            response = client.chat.completions.create(
                model=settings.LLM_MODEL,
                messages=messages,
                tools=TOOL_DEFINITIONS
            )
            choice = response.choices[0]
            tracker.record_llm_call(response.usage)

            if choice.message.tool_calls:
                messages.append(choice.message)
                for tool_call in choice.message.tool_calls:
                    fn_name = tool_call.function.name
                    args = json.loads(tool_call.function.arguments)

                    start_t = time.time()
                    tool_fn = TOOL_MAP.get(fn_name)
                    if tool_fn:
                        res = tool_fn(**args)
                    else:
                        res = {"error": "Tool not found"}
                    duration_ms = (time.time() - start_t) * 1000
                    tracker.record_tool_call(fn_name, duration_ms)

                    messages.append({
                        "role": "tool",
                        "tool_call_id": tool_call.id,
                        "content": json.dumps(res)
                    })
            else:
                reply = choice.message.content
                reply = sanitize_output(reply)

                history_manager.append_message(chat_id, "user", text)
                history_manager.append_message(chat_id, "assistant", reply)

                return {
                    "message_id": message.get("message_id"),
                    "chat_id": chat_id,
                    "reply_text": reply,
                    "metadata": tracker.finalize()
                }
    except Exception as e:
        tracker.record_error(str(e))
        return {
            "message_id": message.get("message_id"),
            "chat_id": chat_id,
            "reply_text": "Maaf, terjadi kendala teknis saat memproses pesan Anda.",
            "metadata": tracker.finalize()
        }
