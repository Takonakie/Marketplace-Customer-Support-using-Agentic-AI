import os
from dotenv import load_dotenv
from pydantic_settings import BaseSettings

# Load .env dari root directory jika ada
load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), "..", "..", ".env"))
load_dotenv()

class Settings(BaseSettings):
    DATABASE_URL: str = os.getenv("DATABASE_URL", "postgres://user:password@localhost:5433/agenticsdk?sslmode=disable")
    REDIS_URL: str = os.getenv("REDIS_URL", "redis://localhost:6380")
    LLM_API_KEY: str = os.getenv("LLM_API_KEY", "")
    LLM_BASE_URL: str = os.getenv("LLM_BASE_URL", "https://api.gutsai.id/v1")
    LLM_MODEL: str = os.getenv("LLM_MODEL", "gemini-3.7-flash")
    GO_SERVICE_URL: str = os.getenv("GO_SERVICE_URL", "http://localhost:8080")

settings = Settings()
