import requests

BASE_URL = "http://localhost:9000"
TIMEOUT = 30
HEADERS = {
    "Accept": "application/json"
}

def test_get_treasury_status_and_balance():
    url = f"{BASE_URL}/dao/treasury"
    try:
        response = requests.get(url, headers=HEADERS, timeout=TIMEOUT)
        # Assert response code is 200 OK
        assert response.status_code == 200, f"Expected status code 200 but got {response.status_code}"
        data = response.json()
        # Expected key in treasury status and balance
        expected_keys = {"balance"}

        # Assert all expected keys are present in response
        missing_keys = expected_keys - data.keys()
        assert not missing_keys, f"Missing keys in treasury response: {missing_keys}"

        # Validate types and values (basic sanity checks)
        assert (isinstance(data["balance"], (int, float)) and data["balance"] >= 0), "Invalid treasury balance"

    except requests.RequestException as e:
        assert False, f"Request to {url} failed with exception: {e}"

test_get_treasury_status_and_balance()
