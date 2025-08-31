import requests

BASE_URL = "http://localhost:9000"
TIMEOUT = 30

def test_get_transaction_by_hash():
    headers = {
        "Accept": "application/json"
    }

    # Use a valid 64-char hex string as a valid transaction hash placeholder
    valid_tx_hash = "a" * 64

    # 1) Test valid transaction hash format
    url_valid = f"{BASE_URL}/tx/{valid_tx_hash}"
    resp_valid = requests.get(url_valid, headers=headers, timeout=TIMEOUT)
    # Expect 200 OK or 404 Not Found
    assert resp_valid.status_code in {200, 404}, f"Expected status 200 or 404 for valid tx hash, got {resp_valid.status_code}"

    if resp_valid.status_code == 200:
        data_valid = resp_valid.json()
        # Basic validations on returned transaction info
        assert isinstance(data_valid, dict), "Transaction info should be a JSON object"
        # Due to PRD no guarantee of 'hash' field, only assert it's a string if present
        if "hash" in data_valid:
            assert isinstance(data_valid["hash"], str), "Returned hash field should be a string"

        # Additional fields presence (based on typical transaction, no full schema provided)
        assert any(k in data_valid for k in ["block_hash", "from", "to", "amount"]), "Expected transaction fields missing"

    # 2) Test invalid transaction hash (bad format)
    invalid_hash = "!!!invalidhash@@@"
    url_invalid = f"{BASE_URL}/tx/{invalid_hash}"
    resp_invalid = requests.get(url_invalid, headers=headers, timeout=TIMEOUT)
    # Expect client error status for invalid tx hash
    assert resp_invalid.status_code in {400, 404, 422}, f"Expected client error status for invalid tx hash, got {resp_invalid.status_code}"

    # 3) Test non-existent but well-formed transaction hash
    non_existent_hash = "f" * 64
    url_nonexist = f"{BASE_URL}/tx/{non_existent_hash}"
    resp_nonexist = requests.get(url_nonexist, headers=headers, timeout=TIMEOUT)
    # Expect 404 Not Found or similar indicating no transaction found
    assert resp_nonexist.status_code == 404, f"Expected 404 for non-existent tx hash, got {resp_nonexist.status_code}"


test_get_transaction_by_hash()
