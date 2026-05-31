from pydantic import BaseModel
from typing import List

class ChatMessage(BaseModel):
    """A single turn in the conversation; role is 'user' or 'assistant'."""

    role: str  # user or assistant
    content: str

class ChatRequest(BaseModel):
    """Request body for the streaming chat endpoint."""

    message: str
    conversation_history: List[ChatMessage] = []
    user_latitude: float
    user_longitude: float

class ChatStreamChunk(BaseModel):
    """A single SSE payload chunk returned by the streaming chat endpoint."""

    text: str
