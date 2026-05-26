from google import genai
from app.config import settings
import json


class GeminiService:
    def __init__(self):
        self.client = genai.Client(api_key=settings.GEMINI_API_KEY)
        self.model = settings.GEMINI_MODEL

    async def extract_intent(
        self,
        user_message: str,
        user_lat: float,
        user_lng: float,
        conversation_history: list
    ) -> dict:
        """Extract search intent from user message using Gemini."""

        system_prompt = f"""You are a location assistant. The user is at coordinates {user_lat:.5f}, {user_lng:.5f}.
Extract their search intent and respond ONLY with a JSON object — no markdown, no explanation:
{{
  "place_type": "<Google Places API type: restaurant|cafe|mosque|park|hospital|pharmacy|supermarket|bank|atm|bakery|bar|night_club|tourist_attraction|museum|school|gas_station>",
  "keyword": "<specific keyword to refine search, e.g. 'kebap', 'lahmacun', 'döner', or empty string>",
  "radius_meters": <number 300-5000>,
  "response_language": "<ISO 639-1 language code the user wrote in>"
}}

Rules:
- For kebap/döner/lahmacun/baklava etc → place_type: "restaurant", keyword: the food name
- For broad requests like "restaurants near me" → keyword: ""
- Larger radius if user says "around here", smaller if "very close" or "walking distance"
- Default radius: 1000
- Detect language from user input (en, tr, ar, etc.)"""

        messages = []
        for msg in conversation_history[-6:]:
            messages.append(f"{msg['role']}: {msg['content']}")
        messages.append(f"user: {user_message}")

        full_prompt = f"{system_prompt}\n\nConversation:\n" + "\n".join(messages)

        response = await self.client.aio.models.generate_content(
            model=self.model,
            contents=full_prompt,
        )

        try:
            text = response.text.strip()
            if text.startswith("```"):
                text = text.split("```")[1]
                if text.startswith("json"):
                    text = text[4:]
            intent = json.loads(text.strip())
            return intent
        except json.JSONDecodeError:
            return {
                "place_type": "restaurant",
                "keyword": "",
                "radius_meters": 1000,
                "response_language": "en"
            }

    async def stream_chat_response(
        self,
        user_message: str,
        places: list,
        intent: dict,
        user_lat: float,
        user_lng: float,
        conversation_history: list
    ):
        """Generate streaming chat response using Gemini."""

        if not places:
            places_context = "No places were found nearby for this search."
        else:
            places_list = []
            for i, place in enumerate(places[:5], 1):
                distance = self._format_distance(place.get("distance_meters", 0))
                rating_text = ""
                if place.get("rating"):
                    rating_text = f", rated {place['rating']}/5"

                open_text = ""
                if place.get("is_open_now") is not None:
                    open_text = ", open now" if place["is_open_now"] else ", currently closed"

                places_list.append(
                    f"{i}. {place['name']} — {distance} away{rating_text}{open_text}\n"
                    f"   Address: {place['vicinity']}"
                )
            places_context = "\n".join(places_list)

        lang = intent.get("response_language", "en")

        system_prompt = f"""You are a friendly local guide assistant. The user is at {user_lat:.4f}, {user_lng:.4f}.
You have just searched nearby places and found the following results:

{places_context}

Respond in {lang} language. Be warm, concise, and helpful.
- Briefly introduce what you found (1 sentence)
- Mention the top 2-3 highlights naturally (closest, highest rated, open now)
- End with a helpful tip or question to refine further
- Do NOT list all results verbatim — the UI cards already show them
- Keep response under 120 words"""

        messages = []
        for msg in conversation_history[-6:]:
            messages.append(f"{msg['role']}: {msg['content']}")
        messages.append(f"user: {user_message}")

        full_prompt = f"{system_prompt}\n\nConversation:\n" + "\n".join(messages)

        async for chunk in await self.client.aio.models.generate_content_stream(
            model=self.model,
            contents=full_prompt,
        ):
            if chunk.text:
                yield chunk.text

    def _format_distance(self, meters: int) -> str:
        if meters >= 1000:
            return f"{meters / 1000:.1f} km"
        return f"{meters} m"


gemini_service = GeminiService()
