from datetime import datetime

import httpx

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
    return datetime.utcnow().timestamp() > expires_at
