import httpx
from app.config import settings
from app.rag.retriever import search_docs

def query_orders(customer_id: str) -> dict:
    url = f"{settings.GO_SERVICE_URL}/api/orders?customer_id={customer_id}"
    resp = httpx.get(url, timeout=5.0)
    return resp.json()

def get_purchase_total(customer_id: str) -> dict:
    url = f"{settings.GO_SERVICE_URL}/api/orders/total?customer_id={customer_id}"
    resp = httpx.get(url, timeout=5.0)
    return resp.json()

def create_case(name: str, criticality: str, description: str, division: str) -> dict:
    url = f"{settings.GO_SERVICE_URL}/api/cases"
    payload = {
        "name": name,
        "criticality": criticality,
        "description": description,
        "assign_to_division": division
    }
    resp = httpx.post(url, json=payload, timeout=5.0)
    return resp.json()

def search_sop(query: str) -> str:
    docs = search_docs(query)
    if not docs:
        return "Tidak ditemukan SOP yang relevan."
    return "\n\n".join([f"--- {d['name']} ---\n{d['doc']}" for d in docs])

TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": "query_orders",
            "description": "Mendapatkan daftar pesanan customer berdasarkan customer_id",
            "parameters": {
                "type": "object",
                "properties": {
                    "customer_id": {"type": "string"}
                },
                "required": ["customer_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_purchase_total",
            "description": "Mendapatkan total akumulasi pembelian customer tahun ini",
            "parameters": {
                "type": "object",
                "properties": {
                    "customer_id": {"type": "string"}
                },
                "required": ["customer_id"]
            }
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
    "create_case": create_case,
    "search_sop": search_sop
}
