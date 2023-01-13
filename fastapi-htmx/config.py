from pydantic import BaseSettings

class Settings(BaseSettings):
    strava_client_id: str
    strava_client_secret: str
    strava_oauth_authorize: str = "https://www.strava.com/oauth/authorize"
    strava_oauth_token: str = "https://www.strava.com/oauth/token"
    session_secret_key: str = "fake-123"

settings = Settings()
