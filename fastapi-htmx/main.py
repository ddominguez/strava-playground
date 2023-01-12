import os
import urllib.parse

from fastapi import FastAPI, Request, status
from fastapi.responses import HTMLResponse, RedirectResponse, Response
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from starlette.middleware.sessions import SessionMiddleware
import httpx

# TODO: Move this to pydantic settings config
STRAVA_CLIENT_ID = os.environ.get("STRAVA_CLIENT_ID")
STRAVA_CLIENT_SECRET = os.environ.get("STRAVA_CLIENT_SECRET")
STRAVA_OAUTH_AUTHORIZE = "https://www.strava.com/oauth/authorize"
STRAVA_OAUTH_TOKEN = "https://www.strava.com/oauth/token"


app = FastAPI()
app.mount("/static", StaticFiles(directory="static"), name="static")

# TODO: Move secret key to env variable and settings config 
app.add_middleware(SessionMiddleware, secret_key="change-me-123")

templates = Jinja2Templates(directory="templates")

@app.get("/", response_class=HTMLResponse)
async def index(request: Request):
    strava_user = request.session.get("strava_user")
    if not strava_user:
        return templates.TemplateResponse("strava_login.html", {"request": request})

    response = httpx.get(
        "https://www.strava.com/api/v3/activities",
        params={
            "per_page": 10
        },
        headers={
            "Authorization": f"Bearer {strava_user.get('access_token')}"
        }
    )
    response.raise_for_status()
    activities = response.json()
    request.state.activities = activities
    return templates.TemplateResponse(
        "content.html",
        {"request": request, "activities": activities}
    )

@app.get("/strava_authorize")
async def strava_authorize(request: Request):
    params = {
        "client_id": STRAVA_CLIENT_ID,
        "redirect_uri": f"{request.base_url}strava_redirect",
        "response_type": "code",
        "scope": "activity:read_all"
    }
    return RedirectResponse(
        f"{STRAVA_OAUTH_AUTHORIZE}?{urllib.parse.urlencode(params)}"
    )

@app.get("/strava_redirect")
async def strava_redirect(request: Request, code: str):
    if not code:
        return Response(
            content="Error: Missing code param",
            status_code=status.HTTP_400_BAD_REQUEST
        )
    response = httpx.post(
        STRAVA_OAUTH_TOKEN,
        data={
            "client_id": STRAVA_CLIENT_ID,
            "client_secret": STRAVA_CLIENT_SECRET,
            "code": code,
            "grant_type": "authorization_code"
        }
    )

    response.raise_for_status()
    request.session["strava_user"] = response.json()
    return RedirectResponse("/")

