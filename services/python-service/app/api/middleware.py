import time
import logging
from fastapi import Request

logger = logging.getLogger(__name__)


async def log_requests(request: Request, call_next):
    """ASGI middleware that logs method, path, status code, and duration for every request."""
    start = time.time()
    response = await call_next(request)
    duration_ms = int((time.time() - start) * 1000)
    logger.info(
        f"{request.method} {request.url.path} "
        f"status={response.status_code} duration={duration_ms}ms"
    )
    return response
