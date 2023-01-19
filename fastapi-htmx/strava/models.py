from datetime import datetime
import math

from pydantic import BaseModel

class ActivityOut(BaseModel):
    id: int
    name: str
    distance: float
    start_date: datetime
    elapsed_time: str
    moving_time: str
    pace: str

    @staticmethod
    def convert_meters_to_miles(meters: float):
        return round(meters * 0.000621371, 2)

    @staticmethod
    def minutes_per_mile_format(time_in_seconds: int, distance_in_meters: int):
       minutes = math.floor(time_in_seconds / 60)
       miles = ActivityOut.convert_meters_to_miles(distance_in_meters)
       pace = minutes / miles
       pace_minutes = math.floor(pace)
       pace_seconds = round((pace - pace_minutes) * 60)
       return f"{pace_minutes}:{pace_seconds}"

    @staticmethod
    def seconds_to_hms(seconds: int):
        h = math.floor(seconds / 3600)
        m = math.floor(seconds % 3600 / 60)
        s = math.floor(seconds % 3600 % 60)
        return f"{h:02}:{m:02}:{s:02}"

    @classmethod
    def build(cls, activity: dict) -> "ActivityOut":
        return cls(
            id=activity["id"],
            name=activity["name"],
            distance=cls.convert_meters_to_miles(activity["distance"]),
            start_date=activity["start_date"],
            elapsed_time=cls.seconds_to_hms(activity["elapsed_time"]),
            moving_time=cls.seconds_to_hms(activity["moving_time"]),
            pace=cls.minutes_per_mile_format(activity["moving_time"], activity["distance"])
        )
