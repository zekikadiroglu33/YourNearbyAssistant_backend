from pydantic import BaseModel
from typing import Optional, List, Dict

class Intent(BaseModel):
    """Structured search intent extracted from a natural-language user query."""

    place_type: str
    keyword: Optional[str] = None
    radius_meters: int = 1000
    response_language: str = "en"

class IntentRequest(BaseModel):
    """Request body for the intent extraction endpoint."""

    user_message: str
    user_latitude: float
    user_longitude: float
    conversation_history: List[Dict[str, str]] = []
