from fastapi import APIRouter, HTTPException
from fastapi.responses import StreamingResponse
from app.models.intent import IntentRequest
from app.models.chat import ChatRequest
from app.services.gemini_service import gemini_service
import httpx
from app.config import settings

router = APIRouter()


@router.post("/api/v1/chat/intent")
async def extract_intent(request: IntentRequest):
    """Extract search intent from user message."""
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
    """Stream AI chat response with places context."""
    try:
        # 1. Extract intent
        intent = await gemini_service.extract_intent(
            user_message=request.message,
            user_lat=request.user_latitude,
            user_lng=request.user_longitude,
            conversation_history=[
                {"role": msg.role, "content": msg.content}
                for msg in request.conversation_history
            ]
        )

        # 2. Fetch places from Go service
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

        # 3. Stream response
        async def generate():
            async for chunk in gemini_service.stream_chat_response(
                user_message=request.message,
                places=places,
                intent=intent,
                user_lat=request.user_latitude,
                user_lng=request.user_longitude,
                conversation_history=[
                    {"role": msg.role, "content": msg.content}
                    for msg in request.conversation_history
                ]
            ):
                yield f"data: {chunk}\n\n"

        return StreamingResponse(generate(), media_type="text/event-stream")

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/health")
async def health_check():
    return {"status": "healthy", "service": "python-ai"}
