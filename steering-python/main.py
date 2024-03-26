import os
from fastapi import FastAPI, Request
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
from dotenv import load_dotenv

from api_endpoints import router as http_router

load_dotenv('/app/.env')


app = FastAPI()
app.include_router(http_router)
origins = [
    "http://localhost",
    "http://0.0.0.0",
    # os.environ.get("FRONTEND_URL"),
]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/health")
async def root():
    return 200

if __name__ == "__main__":
    uvicorn.run("main:app", port=8000, reload=True, debug=True)  # TODO change debug eventually to False
