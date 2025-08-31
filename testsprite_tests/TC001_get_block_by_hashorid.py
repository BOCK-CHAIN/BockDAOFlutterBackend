import requests

BASE_URL = "http://localhost:9000"
TIMEOUT = 30
HEADERS = {
    "Accept": "application/json"
}

def test_get_block_by_hashorid():
    """
    Test GET /block/{hashorid} endpoint for:
    - Valid block height (integer as string)
    - Valid block hash (string)
    - Invalid hashorid (incorrect format)
    - Non-existent hashorid (valid format but no corresponding block)
    """
    # 1. First, get a valid block hash or height by retrieving the genesis block or latest block.
    # Since no direct "/block/latest" in doc, try "0" height assuming Genesis block is 0.
    valid_height = "0"
    valid_block_hash = None

    try:
        # Get block by height 0 (likely genesis block)
        resp = requests.get(f"{BASE_URL}/block/{valid_height}", headers=HEADERS, timeout=TIMEOUT)
        assert resp.status_code == 200, f"Expected 200 for valid height {valid_height}, got {resp.status_code}"
        block_data = resp.json()
        assert isinstance(block_data, dict), "Response is not a JSON object"
        
        # Extract block hash if present in response (heuristic keys)
        # Try common keys: 'hash', 'block_hash', or present top level key that looks like hash
        for key in ["hash", "block_hash", "id"]:
            if key in block_data and isinstance(block_data[key], str):
                valid_block_hash = block_data[key]
                break
        # If no hash found, fallback to first string value in block data with length >= 20 (typical hash length)
        if not valid_block_hash:
            for v in block_data.values():
                if isinstance(v, str) and len(v) >= 20:
                    valid_block_hash = v
                    break

    except Exception as e:
        raise AssertionError(f"Failed to retrieve block by valid height {valid_height}: {e}")

    # 2. Test GET block by valid block hash if found
    if valid_block_hash:
        resp = requests.get(f"{BASE_URL}/block/{valid_block_hash}", headers=HEADERS, timeout=TIMEOUT)
        assert resp.status_code == 200, f"Expected 200 for valid block hash {valid_block_hash}, got {resp.status_code}"
        block_data_hash = resp.json()
        assert isinstance(block_data_hash, dict), "Response for block hash is not a JSON object"
    else:
        # No valid block hash found, log but continue testing other cases
        print("Warning: No valid block hash extracted from block data to test GET by hash.")

    # 3. Test GET block with invalid hashorid (e.g. invalid format string)
    invalid_hashorid = "!!!invalid_hash@@@"
    resp = requests.get(f"{BASE_URL}/block/{invalid_hashorid}", headers=HEADERS, timeout=TIMEOUT)
    # We expect client or server error, commonly 400 or 404
    assert resp.status_code in (400, 404), f"Expected 400 or 404 for invalid hashorid, got {resp.status_code}"

    # 4. Test GET block with non-existent but valid hashorid
    # A valid numeric height string that likely does not exist (e.g. 9999999) or valid hash format but no block
    non_existent_height = "9999999"
    resp = requests.get(f"{BASE_URL}/block/{non_existent_height}", headers=HEADERS, timeout=TIMEOUT)
    # Accept either 400 or 404 for non-existent block
    assert resp.status_code in (400, 404), f"Expected 400 or 404 for non-existent height {non_existent_height}, got {resp.status_code}"

    if valid_block_hash:
        non_existent_hash = valid_block_hash[:-1] + ("0" if valid_block_hash[-1] != "0" else "1")
        resp = requests.get(f"{BASE_URL}/block/{non_existent_hash}", headers=HEADERS, timeout=TIMEOUT)
        assert resp.status_code in (400, 404), f"Expected 400 or 404 for non-existent hash {non_existent_hash}, got {resp.status_code}"

test_get_block_by_hashorid()
