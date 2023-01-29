import urllib.parse

from fastapi import FastAPI, Request, status
from fastapi.responses import HTMLResponse, RedirectResponse, Response
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from starlette.middleware.sessions import SessionMiddleware

from config.settings import settings
from strava import api, models, utils

app = FastAPI(include_in_schema=False)
app.mount("/static", StaticFiles(directory="static"), name="static")
app.add_middleware(
    SessionMiddleware,
    secret_key=settings.session_secret_key,
    max_age=settings.session_max_age,
)

templates = Jinja2Templates(directory="templates")

# TODO: Replace with better cache library???
activity_cache: dict[int, list[models.ActivityOut]] = {}


@app.get("/", response_class=HTMLResponse)
async def index(request: Request):
    user = utils.get_user_from_session(request.session.get("strava_user"))
    if not user:
        return RedirectResponse("/login")
    return RedirectResponse("/activities")


@app.get("/login", response_class=HTMLResponse)
async def login(request: Request):
    user = utils.get_user_from_session(request.session.get("strava_user"))
    if user:
        return RedirectResponse("/activities")
    return templates.TemplateResponse("strava_login.html", {"request": request})


@app.get("/activities", response_class=HTMLResponse)
async def activities(request: Request):
    user = utils.get_user_from_session(request.session.get("strava_user"))
    if not user:
        return RedirectResponse("/login")

    activities: list[models.ActivityOut] = [
        models.ActivityOut.build(activity)
        for activity in api.get_activities(user.get("access_token", ""))
    ]
    activity_cache[user["athlete"]["id"]] = activities
    return templates.TemplateResponse(
        "content.html",
        {"request": request, "activities": activities, "activity": activities[0]},
    )


@app.get("/activities/{activity_id}")
async def get_activity(request: Request, activity_id: int):
    user = utils.get_user_from_session(request.session.get("strava_user"))
    # HTMX will trigger a page refresh if we don't have a strava user.
    if not user:
        return Response(headers={"HX-Refresh": "true"})
    strava_user_activities = activity_cache.get(user["athlete"]["id"])
    activity = None
    if not strava_user_activities:
        return Response(
            content=f"Activity Id {activity_id} not found.",
            status_code=status.HTTP_400_BAD_REQUEST,
        )
    for item in strava_user_activities:
        if item.id == activity_id:
            activity = item
            break
    return templates.TemplateResponse(
        "activity.html", {"request": request, "activity": activity}
    )


@app.get("/strava_authorize")
async def strava_authorize(request: Request):
    params = {
        "client_id": settings.strava_client_id,
        "redirect_uri": f"{request.base_url}strava_redirect",
        "response_type": "code",
        "scope": "activity:read_all",
    }
    return RedirectResponse(
        f"{settings.strava_oauth_authorize}?{urllib.parse.urlencode(params)}"
    )


@app.get("/strava_redirect")
async def strava_redirect(request: Request, code: str):
    if not code:
        return Response(
            content="Error: Missing code param", status_code=status.HTTP_400_BAD_REQUEST
        )
    request.session["strava_user"] = api.authorize_code(code)
    return RedirectResponse("/activities")
