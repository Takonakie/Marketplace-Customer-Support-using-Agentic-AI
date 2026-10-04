import httpx
from app.config import settings
from app.rag.retriever import search_docs

HTTP_TIMEOUT = 10.0

# Field internal yang dibersihkan sebelum dikirim ke LLM demi keamanan data
_INTERNAL_ORDER_FIELDS = {"uuid", "customer_id"}


def _get_json(url: str, params: dict) -> dict | list:
    """HTTP GET helper dengan timeout dan penanganan error terstruktur."""
    try:
        resp = httpx.get(url, params=params, timeout=HTTP_TIMEOUT)
        if resp.status_code >= 400:
            return {"error": f"Layanan data mengembalikan status {resp.status_code}"}
        return resp.json()
    except httpx.HTTPError as e:
        return {"error": f"Layanan data tidak dapat dihubungi: {type(e).__name__}"}
    except ValueError:
        return {"error": "Respons layanan data tidak valid"}


# Keamanan IDOR: customer_id di-inject langsung dari session server, bukan dari LLM

def query_orders(customer_id: str) -> dict | list:
    data = _get_json(f"{settings.GO_SERVICE_URL}/api/orders", {"customer_id": customer_id})
    if isinstance(data, list):
        return [{k: v for k, v in o.items() if k not in _INTERNAL_ORDER_FIELDS} for o in data]
    if data is None:
        return []
    return data


def get_purchase_total(customer_id: str) -> dict:
    data = _get_json(f"{settings.GO_SERVICE_URL}/api/orders/total", {"customer_id": customer_id})
    if isinstance(data, dict):
        data.pop("customer_id", None)
    return data


def create_case(name: str, criticality: str, description: str, division: str, customer_id: str = "") -> dict:
    payload = {
        "name": name,
        "criticality": criticality,
        "description": f"[{customer_id}] {description}" if customer_id else description,
        "assign_to_division": division,
    }
    try:
        resp = httpx.post(f"{settings.GO_SERVICE_URL}/api/cases", json=payload, timeout=HTTP_TIMEOUT)
        if resp.status_code >= 400:
            return {"error": f"Gagal membuat case (status {resp.status_code})"}
        data = resp.json()
        # Hanya kembalikan informasi yang relevan untuk customer (jangan bocorkan UUID staf)
        return {"case_id": data.get("uuid"), "status": data.get("status"), "name": data.get("name")}
    except httpx.HTTPError as e:
        return {"error": f"Layanan case tidak dapat dihubungi: {type(e).__name__}"}
    except ValueError:
        return {"error": "Respons layanan case tidak valid"}


def query_cases(customer_id: str) -> dict | list:
    data = _get_json(f"{settings.GO_SERVICE_URL}/api/cases", {"customer_id": customer_id})
    if isinstance(data, list):
        return [
            {
                "case_id": c.get("uuid"),
                "name": c.get("name"),
                "status": c.get("status"),
                "datetime": c.get("datetime"),
                "description": c.get("description"),
                "resolution": c.get("resolution", "")
            }
            for c in data
        ]
    if data is None:
        return []
    return data


def search_sop(query: str) -> str:
    try:
        docs = search_docs(query)
    except Exception as e:
        return f"Pencarian SOP gagal: {type(e).__name__}"
    if not docs:
        return "Tidak ditemukan SOP yang relevan."
    return "\n\n".join([f"--- {d['name']} ---\n{d['doc']}" for d in docs])


# Tool yang membutuhkan identitas customer -> di-inject oleh core, bukan oleh LLM
CUSTOMER_SCOPED_TOOLS = {"query_orders", "get_purchase_total", "create_case", "query_cases"}

TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": "query_orders",
            "description": "Mendapatkan daftar pesanan milik customer yang sedang chat (identitas customer otomatis dari sistem)",
            "parameters": {"type": "object", "properties": {}, "required": []}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_purchase_total",
            "description": "Mendapatkan total akumulasi pembelian customer yang sedang chat (identitas customer otomatis dari sistem)",
            "parameters": {"type": "object", "properties": {}, "required": []}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "query_cases",
            "description": "Mengecek daftar status tiket komplain/kasus milik customer (misal statusnya open, in_progress, resolved, atau closed)",
            "parameters": {"type": "object", "properties": {}, "required": []}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "create_case",
            "description": "Membuat case/tiket baru untuk tim internal saat komplain/refund tidak bisa diselesaikan langsung",
            "parameters": {
                "type": "object",
                "properties": {
                    "name": {"type": "string", "description": "Judul singkat kasus"},
                    "criticality": {"type": "string", "enum": ["low", "medium", "high", "critical"]},
                    "description": {"type": "string", "description": "Penjelasan detail masalah customer"},
                    "division": {"type": "string", "enum": ["payment", "shipping", "IT"]}
                },
                "required": ["name", "criticality", "description", "division"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "search_sop",
            "description": "Mencari prosedur operasional standar (SOP) terkait pertanyaan/masalah",
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {"type": "string", "description": "Kata kunci pencarian SOP"}
                },
                "required": ["query"]
            }
        }
    }
]

TOOL_MAP = {
    "query_orders": query_orders,
    "get_purchase_total": get_purchase_total,
    "query_cases": query_cases,
    "create_case": create_case,
    "search_sop": search_sop
}
