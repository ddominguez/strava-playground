import urllib.parse

from fastapi import FastAPI, Request, status
from fastapi.responses import HTMLResponse, RedirectResponse, Response
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from starlette.middleware.sessions import SessionMiddleware

from config import settings
import strava

app = FastAPI()
app.mount("/static", StaticFiles(directory="static"), name="static")
app.add_middleware(SessionMiddleware, secret_key=settings.session_secret_key, max_age=settings.session_max_age)

templates = Jinja2Templates(directory="templates")

@app.get("/", response_class=HTMLResponse, include_in_schema=False)
async def index(request: Request):
    strava_user = request.session.get("strava_user")
    # TODO: Refresh token if expired.
    if not strava_user or strava.is_access_token_expired(strava_user.get("expires_at")):
        return templates.TemplateResponse("strava_login.html", {"request": request})

    request.state.activities = strava.get_activities(strava_user.get('access_token'))
    return templates.TemplateResponse(
        "content.html",
        {"request": request}
    )

@app.get("/strava_authorize", include_in_schema=False)
async def strava_authorize(request: Request):
    params = {
        "client_id": settings.strava_client_id,
        "redirect_uri": f"{request.base_url}strava_redirect",
        "response_type": "code",
        "scope": "activity:read_all"
    }
    return RedirectResponse(
        f"{settings.strava_oauth_authorize}?{urllib.parse.urlencode(params)}"
    )

@app.get("/strava_redirect", include_in_schema=False)
async def strava_redirect(request: Request, code: str):
    if not code:
        return Response(
            content="Error: Missing code param",
            status_code=status.HTTP_400_BAD_REQUEST
        )
    request.session["strava_user"] = strava.authorize_code(code)
    return RedirectResponse("/")

