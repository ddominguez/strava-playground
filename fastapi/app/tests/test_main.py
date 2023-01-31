from fastapi.testclient import TestClient
import pytest

from app.main import app

client = TestClient(app)


@pytest.fixture
def mock_get_user_from_session_falsy(monkeypatch):
    monkeypatch.setattr("app.strava.utils.get_user_from_session", lambda _: None)


@pytest.fixture
def mock_get_user_from_session_truthy(monkeypatch):
    monkeypatch.setattr("app.strava.utils.get_user_from_session", lambda _: {"ok": 1})


def test_index_redirects_to_login(mock_get_user_from_session_falsy):
    response = client.get("/", follow_redirects=False)
    assert response.is_redirect
    assert response.headers.get("location") == "/login"


def test_index_redirects_to_activities(mock_get_user_from_session_truthy):
    response = client.get("/", follow_redirects=False)
    assert response.is_redirect
    assert response.headers.get("location") == "/activities"
