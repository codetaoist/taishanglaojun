import os
import json
import pytest
import requests

BASE_URL = os.getenv("BASE_URL", "http://localhost:8080")

try:
    requests.get(BASE_URL, timeout=2)
    SERVER_UP = True
except requests.exceptions.RequestException:
    SERVER_UP = False

ERROR_CODES = {
    "OK",
    "INVALID_ARGUMENT",
    "UNAUTHENTICATED",
    "PERMISSION_DENIED",
    "NOT_FOUND",
    "CONFLICT",
    "FAILED_PRECONDITION",
    "INTERNAL",
    "UNAVAILABLE",
}


def request_json(method: str, path: str, json_body=None, timeout=5):
    url = BASE_URL.rstrip("/") + path
    resp = requests.request(method, url, json=json_body, timeout=timeout)
    # All responses should be unified JSON wrapper
    try:
        data = resp.json()
    except json.JSONDecodeError:
        pytest.fail(f"Non-JSON response for {method} {path}: status={resp.status_code}")
    return resp.status_code, data


def assert_wrapper(payload: dict):
    assert isinstance(payload, dict), "Response payload must be a JSON object"
    assert "code" in payload, "Unified wrapper must include 'code'"
    assert isinstance(payload["code"], str), "'code' must be a string"
    assert payload["code"] in ERROR_CODES, f"Unknown error code: {payload['code']}"
    # Success responses should include 'data'; errors should include 'message'
    if payload["code"] == "OK":
        assert "data" in payload, "Success wrapper must include 'data'"
    else:
        assert "message" in payload, "Error wrapper must include 'message'"


skip_reason = "后端未运行，跳过契约测试（设置 BASE_URL 或启动服务）"


@pytest.mark.skipif(not SERVER_UP, reason=skip_reason)
class TestLaojunPlugins:
    def test_plugins_list_wrapper(self):
        status, body = request_json("GET", "/api/laojun/plugins/list")
        assert_wrapper(body)

    def test_plugins_install_error_wrapper(self):
        # Send invalid payload to trigger error wrapper
        status, body = request_json("POST", "/api/laojun/plugins/install", json_body={})
        assert_wrapper(body)

    def test_plugins_start_error_wrapper(self):
        status, body = request_json("POST", "/api/laojun/plugins/start", json_body={})
        assert_wrapper(body)

    def test_plugins_stop_error_wrapper(self):
        status, body = request_json("POST", "/api/laojun/plugins/stop", json_body={})
        assert_wrapper(body)

    def test_plugins_upgrade_error_wrapper(self):
        status, body = request_json("POST", "/api/laojun/plugins/upgrade", json_body={})
        assert_wrapper(body)

    def test_plugins_uninstall_error_wrapper(self):
        status, body = request_json("POST", "/api/laojun/plugins/uninstall", json_body={})
        assert_wrapper(body)


@pytest.mark.skipif(not SERVER_UP, reason=skip_reason)
class TestTaishangModels:
    def test_models_list_wrapper(self):
        status, body = request_json("GET", "/api/taishang/models")
        assert_wrapper(body)

    def test_models_detail_error_wrapper(self):
        status, body = request_json("GET", "/api/taishang/models/m-unknown")
        assert_wrapper(body)

    def test_models_enable_error_wrapper(self):
        status, body = request_json("POST", "/api/taishang/models/m-unknown/enable")
        assert_wrapper(body)

    def test_models_disable_error_wrapper(self):
        status, body = request_json("POST", "/api/taishang/models/m-unknown/disable")
        assert_wrapper(body)


@pytest.mark.skipif(not SERVER_UP, reason=skip_reason)
class TestTaishangVectors:
    def test_collections_list_wrapper(self):
        status, body = request_json("GET", "/api/taishang/vectors/collections")
        assert_wrapper(body)

    def test_collection_create_error_wrapper(self):
        status, body = request_json(
            "POST",
            "/api/taishang/vectors/collections",
            json_body={"name": "", "dim": 0, "indexType": "INVALID", "metric": "INVALID"},
        )
        assert_wrapper(body)

    def test_upsert_error_wrapper(self):
        status, body = request_json(
            "POST",
            "/api/taishang/vectors/collections/vc-unknown/upsert",
            json_body={"vectors": []},
        )
        assert_wrapper(body)

    def test_query_error_wrapper(self):
        status, body = request_json(
            "POST",
            "/api/taishang/vectors/collections/vc-unknown/query",
            json_body={"topK": 0, "query": {"text": ""}},
        )
        assert_wrapper(body)

    def test_delete_error_wrapper(self):
        status, body = request_json(
            "POST",
            "/api/taishang/vectors/collections/vc-unknown/delete",
            json_body={"ids": ["non-existent"]},
        )
        assert_wrapper(body)


@pytest.mark.skipif(not SERVER_UP, reason=skip_reason)
class TestTaishangTasks:
    def test_tasks_list_wrapper(self):
        status, body = request_json("GET", "/api/taishang/tasks")
        assert_wrapper(body)

    def test_task_submit_error_wrapper(self):
        status, body = request_json(
            "POST",
            "/api/taishang/tasks",
            json_body={"type": "", "payload": {}},
        )
        assert_wrapper(body)

    def test_task_detail_error_wrapper(self):
        status, body = request_json("GET", "/api/taishang/tasks/t-unknown")
        assert_wrapper(body)

    def test_task_cancel_error_wrapper(self):
        status, body = request_json("POST", "/api/taishang/tasks/t-unknown/cancel")
        assert_wrapper(body)

    def test_task_retry_error_wrapper(self):
        status, body = request_json("POST", "/api/taishang/tasks/t-unknown/retry")
        assert_wrapper(body)