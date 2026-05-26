from pydantic import BaseModel
from typing import Optional, List, Dict

class Intent(BaseModel):
    place_type: str
    keyword: Optional[str] = None
    radius_meters: int = 1000
    response_language: str = "en"

class IntentRequest(BaseModel):
    user_message: str
    user_latitude: float
    user_longitude: float
    conversation_history: List[Dict[str, str]] = []
