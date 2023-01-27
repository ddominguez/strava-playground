import httpx

from config.settings import settings


def authorize_code(code: str):
    response = httpx.post(
        settings.strava_oauth_token,
        data={
            "client_id": settings.strava_client_id,
            "client_secret": settings.strava_client_secret,
            "code": code,
            "grant_type": "authorization_code",
        },
    )

    response.raise_for_status()
    return response.json()


def get_activities(access_token: str):
    response = httpx.get(
        "https://www.strava.com/api/v3/activities",
        params={"per_page": 10},
        headers={"Authorization": f"Bearer {access_token}"},
    )
    response.raise_for_status()
    return response.json()
