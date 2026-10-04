import psycopg2
from pgvector.psycopg2 import register_vector
from openai import OpenAI
from app.config import settings

def get_db_conn():
    conn = psycopg2.connect(settings.DATABASE_URL)
    register_vector(conn)
    return conn

def generate_embedding(text: str) -> list[float]:
    # Jika OPENAI_EMBEDDING_API_KEY diisi, gunakan official OpenAI endpoint
    api_key = settings.OPENAI_EMBEDDING_API_KEY or settings.LLM_API_KEY
    if not api_key:
        print("[WARN] API Key untuk embedding tidak diisi, menggunakan dummy zero vector.")
        return [0.0] * 1536

    client_kwargs = {"api_key": api_key}
    if settings.OPENAI_EMBEDDING_BASE_URL:
        client_kwargs["base_url"] = settings.OPENAI_EMBEDDING_BASE_URL
    elif not settings.OPENAI_EMBEDDING_API_KEY and settings.LLM_BASE_URL:
        client_kwargs["base_url"] = settings.LLM_BASE_URL

    client = OpenAI(**client_kwargs)
    try:
        # Jika menggunakan OpenRouter, model default menggunakan prefix 'openai/text-embedding-3-small' atau 'text-embedding-3-small'
        embed_model = "openai/text-embedding-3-small" if "openrouter" in (settings.OPENAI_EMBEDDING_BASE_URL or "").lower() else "text-embedding-3-small"
        response = client.embeddings.create(
            model=embed_model,
            input=text
        )
        return response.data[0].embedding
    except Exception as e:
        print(f"[WARN] Provider embeddings API error ({e}), menggunakan fallback zero vector.")
        return [0.0] * 1536

def load_and_embed_all_docs():
    print(">>> Starting SOP Document Ingestion & Embedding Process <<<")
    conn = get_db_conn()
    cur = conn.cursor()

    # Ambil semua dokumen yang belum punya embedding
    cur.execute("SELECT uuid, name, doc FROM docs WHERE embedding IS NULL")
    rows = cur.fetchall()

    if not rows:
        print("Semua dokumen SOP sudah memiliki embedding. Tidak ada yang diproses.")
        cur.close()
        conn.close()
        return

    print(f"Ditemukan {len(rows)} dokumen SOP yang memerlukan embedding...")

    for doc_uuid, name, doc_content in rows:
        print(f"Processing embedding for SOP: '{name}' (UUID: {doc_uuid})...")
        try:
            vector = generate_embedding(doc_content)
            cur.execute("UPDATE docs SET embedding = %s, updated_at = NOW() WHERE uuid = %s", (vector, doc_uuid))
            conn.commit()
            print(f"✅ Successfully embedded: '{name}'")
        except Exception as e:
            conn.rollback()
            print(f"❌ Failed to embed '{name}': {e}")

    cur.close()
    conn.close()
    print(">>> Ingestion & Embedding Process Completed! <<<")

if __name__ == "__main__":
    load_and_embed_all_docs()
