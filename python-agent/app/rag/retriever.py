import psycopg2
from pgvector.psycopg2 import register_vector
from openai import OpenAI
from app.config import settings

def get_db_conn():
    conn = psycopg2.connect(settings.DATABASE_URL)
    register_vector(conn)
    return conn

def embed_text(text: str) -> list[float]:
    # Jika OPENAI_EMBEDDING_API_KEY diisi, gunakan official OpenAI endpoint
    api_key = settings.OPENAI_EMBEDDING_API_KEY or settings.LLM_API_KEY
    if not api_key:
        # Mock embedding for test/dev mode if key is absent
        return [0.0] * 1536

    client_kwargs = {"api_key": api_key}
    if settings.OPENAI_EMBEDDING_BASE_URL:
        client_kwargs["base_url"] = settings.OPENAI_EMBEDDING_BASE_URL
    elif not settings.OPENAI_EMBEDDING_API_KEY and settings.LLM_BASE_URL:
        client_kwargs["base_url"] = settings.LLM_BASE_URL

    client = OpenAI(**client_kwargs)
    try:
        embed_model = "openai/text-embedding-3-small" if "openrouter" in (settings.OPENAI_EMBEDDING_BASE_URL or "").lower() else "text-embedding-3-small"
        response = client.embeddings.create(
            model=embed_model,
            input=text
        )
        return response.data[0].embedding
    except Exception as e:
        print(f"[WARN] Provider embeddings API error ({e}), menggunakan fallback zero vector.")
        return [0.0] * 1536

def search_docs(query: str, top_k: int = 3) -> list[dict]:
    try:
        query_vector = embed_text(query)
        conn = get_db_conn()
        cur = conn.cursor()
        cur.execute("""
            SELECT uuid, name, doc, 1 - (embedding <=> %s::vector) AS similarity
            FROM docs
            WHERE embedding IS NOT NULL
            ORDER BY embedding <=> %s::vector
            LIMIT %s
        """, (query_vector, query_vector, top_k))
        rows = cur.fetchall()
        cur.close()
        conn.close()

        results = []
        for r in rows:
            results.append({
                "uuid": r[0],
                "name": r[1],
                "doc": r[2],
                "similarity": float(r[3]) if r[3] else 0.0
            })
        return results
    except Exception as e:
        # Fallback keyword search if vector search fails or embeddings are not loaded yet
        conn = get_db_conn()
        cur = conn.cursor()
        cur.execute("SELECT uuid, name, doc FROM docs WHERE doc ILIKE %s LIMIT %s", (f"%{query}%", top_k))
        rows = cur.fetchall()
        cur.close()
        conn.close()
        return [{"uuid": r[0], "name": r[1], "doc": r[2]} for r in rows]
