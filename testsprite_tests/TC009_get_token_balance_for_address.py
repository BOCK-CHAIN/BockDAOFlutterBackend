import requests

BASE_URL = "http://localhost:9000"
TIMEOUT = 30

def test_get_token_balance_for_address():
    """
    Test the GET /dao/token/balance/{address} endpoint for correctness.
    This test will:
    - Create a new token holder by transferring tokens to a new address
    - Query the balance for that address and validate the expected balance
    - Cleanup if necessary (not applicable here since token transfer is blockchain state)
    """
    # For testing, we need a valid address. We'll create one by transferring tokens to a newly generated address.
    # In real scenario, generating a wallet/address should be done properly; here we simulate a random test address.
    import uuid
    test_address = f"testaddress_{uuid.uuid4().hex[:16]}"  # synthetic test address

    # We need to have tokens in this address to verify balance.
    # For that, transfer tokens from a known address with private_key (we simulate it here).
    # Use a predefined source address and private_key for testing (replace with valid test creds).
    source_private_key = "test_private_key_for_source"  # Placeholder: Replace with valid key for actual testing
    transfer_amount = 100

    headers = {"Content-Type": "application/json"}

    # Step 1: Transfer tokens to the test_address to ensure it has tokens
    transfer_payload = {
        "to": test_address,
        "amount": transfer_amount,
        "private_key": source_private_key
    }
    try:
        transfer_response = requests.post(
            f"{BASE_URL}/dao/token/transfer",
            json=transfer_payload,
            headers=headers,
            timeout=TIMEOUT
        )
        assert transfer_response.status_code == 200, f"Token transfer failed: {transfer_response.text}"

        # Step 2: Query the token balance for the test_address
        balance_response = requests.get(
            f"{BASE_URL}/dao/token/balance/{test_address}",
            timeout=TIMEOUT
        )
        assert balance_response.status_code == 200, f"Balance fetch failed: {balance_response.text}"

        balance_data = balance_response.json()
        # The API response structure is unknown, but expect a field named 'balance' or similar
        assert "balance" in balance_data, f"Response missing 'balance' field: {balance_data}"
        balance = balance_data["balance"]
        assert isinstance(balance, (int, float)), f"Balance field is not a number: {balance}"

        # Validate that balance is at least the amount transferred
        assert balance >= transfer_amount, f"Balance {balance} less than transferred amount {transfer_amount}"

    except requests.exceptions.RequestException as e:
        assert False, f"Request failed: {str(e)}"

test_get_token_balance_for_address()