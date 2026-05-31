from fastapi import APIRouter, HTTPException
from fastapi.responses import JSONResponse, StreamingResponse
from app.models.intent import IntentRequest
from app.models.chat import ChatRequest, IntentConfirmationResponse
from app.services.gemini_service import gemini_service
import httpx
from app.config import settings

router = APIRouter()


def _build_confirmation_message(intent: dict) -> str:
    """Build a human-readable confirmation string shown to the user before the search runs."""
    place_type = intent.get("place_type", "places").replace("_", " ")
    keyword = intent.get("keyword", "")
    radius = intent.get("radius_meters", 1000)
    walk_min = max(1, round(radius / 80))

    search_desc = f'"{keyword}" {place_type}' if keyword else place_type
    return (
        f"I'll search for {search_desc} within {radius} m "
        f"(~{walk_min} min walk). Does that sound right?"
    )


@router.post("/api/v1/chat/intent")
async def extract_intent(request: IntentRequest):
    """Parse a natural-language query into structured search intent (place type, keyword, radius, language)."""
    try:
        intent = await gemini_service.extract_intent(
            user_message=request.user_message,
            user_lat=request.user_latitude,
            user_lng=request.user_longitude,
            conversation_history=request.conversation_history
        )
        return intent
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/api/v1/chat/stream")
async def stream_chat(request: ChatRequest):
    """Human-in-the-loop streaming chat.

    Two-step flow:
    1. First call (confirmed_intent omitted): extracts search intent from the user message
       and returns an IntentConfirmationResponse (HTTP 200 JSON) for the user to review.
       The client shows the confirmation_message and lets the user approve or edit intent fields.
    2. Second call (confirmed_intent provided): uses the approved intent to query the Go
       Places service, then streams a Gemini-generated conversational response as SSE.
    """
    try:
        history = [
            {"role": msg.role, "content": msg.content}
            for msg in request.conversation_history
        ]

        # ── Step 1: no confirmed intent yet — extract and ask the user ──────────
        if request.confirmed_intent is None:
            intent = await gemini_service.extract_intent(
                user_message=request.message,
                user_lat=request.user_latitude,
                user_lng=request.user_longitude,
                conversation_history=history
            )
            confirmation = IntentConfirmationResponse(
                intent=intent,
                confirmation_message=_build_confirmation_message(intent)
            )
            return JSONResponse(content=confirmation.model_dump())

        # ── Step 2: user confirmed — run the search and stream ──────────────────
        intent = request.confirmed_intent

        async with httpx.AsyncClient(timeout=10.0) as client:
            places_response = await client.post(
                f"{settings.GO_SERVICE_URL}/api/v1/places/nearby",
                json={
                    "latitude": request.user_latitude,
                    "longitude": request.user_longitude,
                    "radius": intent.get("radius_meters", 1000),
                    "type": intent.get("place_type") or None,
                    "keyword": intent.get("keyword") or None
                }
            )
            places = places_response.json() if places_response.status_code == 200 else []

        async def generate():
            async for chunk in gemini_service.stream_chat_response(
                user_message=request.message,
                places=places,
                intent=intent,
                user_lat=request.user_latitude,
                user_lng=request.user_longitude,
                conversation_history=history
            ):
                yield f"data: {chunk}\n\n"

        return StreamingResponse(generate(), media_type="text/event-stream")

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/health")
async def health_check():
    """Return a simple liveness payload confirming the Python AI service is running."""
    return {"status": "healthy", "service": "python-ai"}
