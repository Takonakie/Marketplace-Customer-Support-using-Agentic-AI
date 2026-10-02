import psycopg2
from pgvector.psycopg2 import register_vector
from openai import OpenAI
from app.config import settings

def get_db_conn():
    conn = psycopg2.connect(settings.DATABASE_URL)
    register_vector(conn)
    return conn

def generate_embedding(text: str) -> list[float]:
    if not settings.LLM_API_KEY:
        print("[WARN] LLM_API_KEY tidak diisi, menggunakan dummy zero vector.")
        return [0.0] * 1536

    client_kwargs = {"api_key": settings.LLM_API_KEY}
    if settings.LLM_BASE_URL:
        client_kwargs["base_url"] = settings.LLM_BASE_URL

    client = OpenAI(**client_kwargs)
    response = client.embeddings.create(
        model="text-embedding-3-small",
        input=text
    )
    return response.data[0].embedding

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
