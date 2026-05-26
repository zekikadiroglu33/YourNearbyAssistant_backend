from pydantic import BaseModel
from typing import List

class ChatMessage(BaseModel):
    role: str  # user or assistant
    content: str

class ChatRequest(BaseModel):
    message: str
    conversation_history: List[ChatMessage] = []
    user_latitude: float
    user_longitude: float

class ChatStreamChunk(BaseModel):
    text: str
