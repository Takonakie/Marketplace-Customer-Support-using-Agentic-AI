import json
import time
import asyncio
import inspect
from openai import AsyncOpenAI
from app.config import settings
from app.agent.prompts import build_system_prompt
from app.agent.tools import TOOL_DEFINITIONS, TOOL_MAP, CUSTOMER_SCOPED_TOOLS
from app.agent.guardrails import validate_input, sanitize_output
from app.monitoring.tracker import MonitoringTracker
from app.chat_history.manager import ChatHistoryManager

async def handle_message(message: dict, history_manager: ChatHistoryManager) -> dict:
    chat_id = message["chat_id"]
    customer_name = message.get("customer_name", "Customer")
    text = message.get("text", "")
    tracker = MonitoringTracker(message.get("message_id", ""), user_message=text)

    guard_err = validate_input(text)
    if guard_err:
        tracker.set_ai_response(guard_err)
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
        reply = f"Hai {customer_name}, sistem menerima pesan Anda: '{text}'. (Dev Mode - API Key belum diisi)"
        history_manager.append_message(chat_id, "user", text)
        history_manager.append_message(chat_id, "assistant", reply)
        tracker.set_ai_response(reply)
        return {
            "message_id": message.get("message_id"),
            "chat_id": chat_id,
            "reply_text": reply,
            "metadata": tracker.finalize()
        }

    async def call_llm(messages_payload):
        # Async call to Primary LLM with automatic failover to Fallback LLM
        primary_kwargs = {"api_key": settings.LLM_API_KEY}
        if settings.LLM_BASE_URL:
            primary_kwargs["base_url"] = settings.LLM_BASE_URL
        
        try:
            client = AsyncOpenAI(**primary_kwargs)
            return await client.chat.completions.create(
                model=settings.LLM_MODEL,
                messages=messages_payload,
                tools=TOOL_DEFINITIONS,
                max_tokens=settings.LLM_MAX_TOKENS
            )
        except Exception as primary_err:
            print(f"[WARN] Primary LLM failed ({primary_err}). Switching to Fallback LLM provider (OpenRouter)...", flush=True)
            if not settings.FALLBACK_LLM_API_KEY:
                raise primary_err
            
            fallback_kwargs = {"api_key": settings.FALLBACK_LLM_API_KEY}
            if settings.FALLBACK_LLM_BASE_URL:
                fallback_kwargs["base_url"] = settings.FALLBACK_LLM_BASE_URL

            fallback_client = AsyncOpenAI(**fallback_kwargs)
            return await fallback_client.chat.completions.create(
                model=settings.FALLBACK_LLM_MODEL,
                messages=messages_payload,
                tools=TOOL_DEFINITIONS,
                max_tokens=settings.LLM_MAX_TOKENS
            )

    try:
        while True:
            response = await call_llm(messages)
            choice = response.choices[0]
            tracker.record_llm_call(response.usage)

            if choice.message.tool_calls:
                msg_dict = {"role": "assistant", "content": choice.message.content or ""}
                msg_dict["tool_calls"] = [
                    {
                        "id": tc.id,
                        "type": tc.type,
                        "function": {"name": tc.function.name, "arguments": tc.function.arguments}
                    } for tc in choice.message.tool_calls
                ]
                messages.append(msg_dict)

                async def execute_single_tool(tool_call):
                    fn_name = tool_call.function.name
                    args = json.loads(tool_call.function.arguments)
                    start_t = time.time()
                    tool_fn = TOOL_MAP.get(fn_name)
                    if tool_fn:
                        if fn_name in CUSTOMER_SCOPED_TOOLS:
                            args["customer_id"] = f"tg_{chat_id}"
                        
                        sig = inspect.signature(tool_fn)
                        valid_args = {k: v for k, v in args.items() if k in sig.parameters}
                        if asyncio.iscoroutinefunction(tool_fn):
                            res = await tool_fn(**valid_args)
                        else:
                            res = await asyncio.to_thread(tool_fn, **valid_args)
                    else:
                        res = {"error": "Tool not found"}
                    
                    duration_ms = (time.time() - start_t) * 1000
                    tracker.record_tool_call(fn_name, duration_ms)
                    return {
                        "role": "tool",
                        "tool_call_id": tool_call.id,
                        "content": json.dumps(res)
                    }

                # Parallel tool execution via asyncio.gather
                tool_results = await asyncio.gather(
                    *[execute_single_tool(tc) for tc in choice.message.tool_calls]
                )
                for res_msg in tool_results:
                    messages.append(res_msg)
            else:
                reply = choice.message.content
                reply = sanitize_output(reply)

                history_manager.append_message(chat_id, "user", text)
                history_manager.append_message(chat_id, "assistant", reply)

                tracker.set_ai_response(reply)
                return {
                    "message_id": message.get("message_id"),
                    "chat_id": chat_id,
                    "reply_text": reply,
                    "metadata": tracker.finalize()
                }
    except Exception as e:
        print(f"[ERROR] Agent handle_message exception: {e}", flush=True)
        err_msg = "Maaf, terjadi kendala teknis saat memproses pesan Anda."
        tracker.record_error(str(e))
        tracker.set_ai_response(err_msg)
        return {
            "message_id": message.get("message_id"),
            "chat_id": chat_id,
            "reply_text": err_msg,
            "metadata": tracker.finalize()
        }
