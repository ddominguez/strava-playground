import json
from base64 import b64encode
from datetime import datetime, timezone

from fastapi.testclient import TestClient
from itsdangerous import TimestampSigner

from app.config.settings import settings
from app.main import app

client = TestClient(app)


# https://github.com/tiangolo/fastapi/issues/929#issuecomment-940982932
def create_session_cookie(data):
    signer = TimestampSigner(str(settings.session_secret_key))
    return signer.sign(b64encode(json.dumps(data).encode("utf-8"))).decode("utf-8")


def test_index_redirects_to_login():
    response = client.get("/", follow_redirects=False)
    assert response.is_redirect
    assert response.headers.get("location") == "/login"


def test_index_redirects_to_activities():
    session_data = {
        "strava_user": {"expires_at": datetime.now(tz=timezone.utc).timestamp() + 100}
    }
    client.cookies = {"session": create_session_cookie(session_data)}
    response = client.get("/", follow_redirects=False)
    assert response.is_redirect
    assert response.headers.get("location") == "/activities"
