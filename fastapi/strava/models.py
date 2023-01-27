from datetime import datetime

from pydantic import BaseModel

from .utils import (
    convert_meters_to_miles,
    minutes_per_mile_format,
    seconds_to_hms,
)


class ActivityOut(BaseModel):
    id: int
    name: str
    distance: float
    start_date: datetime
    elapsed_time: str
    moving_time: str
    pace: str

    @classmethod
    def build(cls, activity: dict) -> "ActivityOut":
        return cls(
            id=activity["id"],
            name=activity["name"],
            distance=convert_meters_to_miles(activity["distance"]),
            start_date=activity["start_date"],
            elapsed_time=seconds_to_hms(activity["elapsed_time"]),
            moving_time=seconds_to_hms(activity["moving_time"]),
            pace=minutes_per_mile_format(activity["moving_time"], activity["distance"]),
        )
