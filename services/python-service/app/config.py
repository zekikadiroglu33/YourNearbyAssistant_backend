import os
from dotenv import load_dotenv

load_dotenv()

class Settings:
    GEMINI_API_KEY: str = os.getenv("GEMINI_API_KEY", "")
    GEMINI_MODEL: str = os.getenv("GEMINI_MODEL", "gemini-2.5-flash")
    GO_SERVICE_URL: str = os.getenv("GO_SERVICE_URL", "http://localhost:8080")
    PORT: int = int(os.getenv("PYTHON_SERVICE_PORT", "8001"))

settings = Settings()
