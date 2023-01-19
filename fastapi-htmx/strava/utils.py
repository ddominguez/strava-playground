from datetime import datetime, timezone
import typing

import httpx

from config import settings

def authorize_code(code: str):
    response = httpx.post(
        settings.strava_oauth_token,
        data={
            "client_id": settings.strava_client_id,
            "client_secret": settings.strava_client_secret,
            "code": code,
            "grant_type": "authorization_code"
        }
    )

    response.raise_for_status()
    return response.json()

def get_activities(access_token: str):
    response = httpx.get(
        "https://www.strava.com/api/v3/activities",
        params={
            "per_page": 10
        },
        headers={
            "Authorization": f"Bearer {access_token}"
        }
    )
    response.raise_for_status()
    return response.json()

def is_access_token_expired(expires_at: int) -> bool:
    return datetime.now(tz=timezone.utc) > datetime.fromtimestamp(expires_at, tz=timezone.utc)

def get_user_from_session(strava_user: dict[str, typing.Any] | None) -> dict[str, typing.Any] | None:
    if not strava_user or is_access_token_expired(strava_user.get("expires_at", 0)):
        return None
    return strava_user

