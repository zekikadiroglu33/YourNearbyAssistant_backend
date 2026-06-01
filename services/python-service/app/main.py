from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from starlette.middleware.base import BaseHTTPMiddleware
from app.api import routes as route_module
from app.api.routes import router
from app.api.middleware import log_requests
from app.config import settings
import httpx
import uvicorn


@asynccontextmanager
async def lifespan(app: FastAPI):
    route_module.http_client = httpx.AsyncClient(timeout=10.0)
    yield
    await route_module.http_client.aclose()


app = FastAPI(title="AI Service", version="1.0.0", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.add_middleware(BaseHTTPMiddleware, dispatch=log_requests)

app.include_router(router)

if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=settings.PORT,
        reload=True
    )
