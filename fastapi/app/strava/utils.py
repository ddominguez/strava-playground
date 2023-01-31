from datetime import datetime, timezone
import math
import typing


def is_access_token_expired(expires_at: int) -> bool:
    return datetime.now(tz=timezone.utc) > datetime.fromtimestamp(
        expires_at, tz=timezone.utc
    )


def get_user_from_session(
    strava_user: dict[str, typing.Any] | None
) -> dict[str, typing.Any] | None:
    if not strava_user or is_access_token_expired(strava_user.get("expires_at", 0)):
        return None
    return strava_user


def convert_meters_to_miles(meters: float):
    return round(meters * 0.000621371, 2)


def minutes_per_mile_format(time_in_seconds: int, distance_in_meters: int):
    minutes = math.floor(time_in_seconds / 60)
    miles = convert_meters_to_miles(distance_in_meters)
    pace = minutes / miles
    pace_minutes = math.floor(pace)
    pace_seconds = round((pace - pace_minutes) * 60)
    return f"{pace_minutes}:{pace_seconds}"


def seconds_to_hms(seconds: int):
    h = math.floor(seconds / 3600)
    m = math.floor(seconds % 3600 / 60)
    s = math.floor(seconds % 3600 % 60)
    return f"{h:02}:{m:02}:{s:02}"
