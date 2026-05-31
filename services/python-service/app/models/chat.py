from pydantic import BaseModel
from typing import Dict, List, Optional, Any

class ChatMessage(BaseModel):
    """A single turn in the conversation; role is 'user' or 'assistant'."""

    role: str  # user or assistant
    content: str

class ChatRequest(BaseModel):
    """Request body for the streaming chat endpoint.

    HITL flow:
    - First call: omit confirmed_intent → backend returns IntentConfirmationResponse for user review.
    - Second call: include confirmed_intent (the approved/edited intent dict) → backend runs the
      Places search and streams the AI response.
    """

    message: str
    conversation_history: List[ChatMessage] = []
    user_latitude: float
    user_longitude: float
    confirmed_intent: Optional[Dict[str, Any]] = None

class IntentConfirmationResponse(BaseModel):
    """Returned by /chat/stream when no confirmed_intent is provided.
    The client should display confirmation_message and let the user approve or edit
    the intent fields before re-submitting with confirmed_intent populated.
    """

    awaiting_confirmation: bool = True
    intent: Dict[str, Any]
    confirmation_message: str

class ChatStreamChunk(BaseModel):
    """A single SSE payload chunk returned by the streaming chat endpoint."""

    text: str
