from google import genai
from app.config import settings
import json


class GeminiService:
    """Wrapper around the Google Gemini API for intent extraction and streaming chat responses."""

    def __init__(self):
        """Initialise the Gemini client and pin the model from settings."""
        self.client = genai.Client(api_key=settings.GEMINI_API_KEY)
        self.model = settings.GEMINI_MODEL

    async def extract_intent(
        self,
        user_message: str,
        user_lat: float,
        user_lng: float,
        conversation_history: list
    ) -> dict:
        """Call Gemini with a structured system prompt to extract place_type, keyword,
        radius_meters, and response_language from the user's natural-language message.
        Falls back to safe defaults if the model returns unparseable JSON."""

        system_prompt = f"""You are a smart travel companion helping someone explore their current surroundings on foot.
The user is standing at coordinates {user_lat:.5f}, {user_lng:.5f}.

Analyse the conversation and extract their search intent. Reply ONLY with a valid JSON object — no markdown, no explanation, no extra text:
{{
  "place_type": "<single most relevant Google Places type: restaurant|cafe|mosque|church|park|hospital|pharmacy|supermarket|convenience_store|bank|atm|bakery|bar|night_club|tourist_attraction|museum|hotel|lodging|gas_station|parking|transit_station|bus_station|subway_station|school|university|library|gym|spa|clothing_store|shopping_mall|electronics_store>",
  "keyword": "<specific term to narrow results, e.g. 'sushi', 'vegan', 'breakfast', 'döner', 'Starbucks', or '' for a broad category search>",
  "radius_meters": <integer between 300 and 5000>,
  "response_language": "<ISO 639-1 code of the language the user wrote in>"
}}

Radius — think in walking minutes (avg ~80 m/min):
- "right next to me", "closest", "I can walk" → 300–500 m (under 7 min)
- no qualifier, general intent → 1000 m (~12 min)
- "nearby", "around here" → 1500 m (~20 min)
- "in this neighbourhood", "this area" → 2000–3000 m
- "anywhere in the city", "I don't mind travelling" → 4000–5000 m
- urgency words ("urgent", "emergency", "I need now") → always 500 m or less

Keyword rules:
- Specific food or dish (kebap, latte, sushi, taco, baklava) → keyword = that item
- Brand or venue name (Starbucks, McDonald's, Marriott) → keyword = that name
- Quality/vibe signal ("cheap eats", "rooftop bar", "cosy café") → keyword = the signal phrase
- Broad category ("restaurants", "parks", "ATMs") → keyword = ""

Language: detect from the user's own words (en, tr, ar, fr, es, de, etc.)"""

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
        """Yield Gemini response text chunks as an async generator.
        Builds context from the top-5 nearby places and the last 6 conversation turns,
        instructing the model to respond in the detected language under 120 words."""

        if not places:
            places_context = "No places were found for this search. The area may have nothing matching, or the radius is too small."
        else:
            places_list = []
            price_symbols = {0: "free", 1: "$", 2: "$$", 3: "$$$", 4: "$$$$"}
            for i, place in enumerate(places[:7], 1):
                meters = place.get("distance_meters", 0)
                walk_min = max(1, round(meters / 80))
                dist_text = f"{self._format_distance(meters)} · ~{walk_min} min walk"

                rating = place.get("rating")
                rating_text = f"rated {rating}/5 ({place.get('user_ratings_total', 0)} reviews)" if rating else "no rating yet"

                is_open = place.get("is_open_now")
                open_text = "open now" if is_open is True else ("closed" if is_open is False else "hours unknown")

                price = place.get("price_level")
                price_text = f" · {price_symbols[price]}" if price is not None else ""

                places_list.append(
                    f"{i}. {place['name']} [{dist_text}{price_text}]\n"
                    f"   {rating_text} · {open_text}\n"
                    f"   {place['vicinity']}"
                )
            places_context = "\n\n".join(places_list)

        lang = intent.get("response_language", "en")

        system_prompt = f"""You are a knowledgeable local travel guide. The user is currently standing at {user_lat:.5f}, {user_lng:.5f} and is looking for somewhere to go right now.

The results below are already sorted by walking distance — closest first:

{places_context}

Your job:
- Respond in {lang} language
- Open with ONE sentence that captures what you found and how close the best options are
- Naturally highlight the top 2–3 picks: lead with the closest that is open, then the highest-rated, then any standout detail (price, vibe, specialty)
- If a place is closed, say so and pivot to the next best open option
- End with a short practical tip or a follow-up question to refine the search
- Do NOT recite all results line-by-line — the app cards already show them
- Keep the response under 100 words; think concise travel-guide copy, not a list"""

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
        """Return a human-readable distance string (e.g. '350 m' or '1.2 km')."""
        if meters >= 1000:
            return f"{meters / 1000:.1f} km"
        return f"{meters} m"


gemini_service = GeminiService()
