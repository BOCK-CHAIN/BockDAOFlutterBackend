import requests
import time

BASE_URL = "http://localhost:9000"
TIMEOUT = 30
HEADERS = {
    "Content-Type": "application/json",
    "Accept": "application/json"
}

def test_get_all_governance_proposals():
    """
    Test GET /dao/proposals returns a list of all governance proposals with correct data structure and content.
    If no proposals exist, create a test proposal, verify inclusion, then delete it.
    """
    proposal_create_url = f"{BASE_URL}/dao/proposal"
    proposals_url = f"{BASE_URL}/dao/proposals"

    # Sample proposal data for creation
    new_proposal_payload = {
        "title": "Test Proposal - Governance API Validation",
        "description": "Proposal created for testing GET /dao/proposals endpoint",
        "proposal_type": "general",
        "voting_type": "token-based",
        "duration": 3600,      # 1 hour
        "threshold": 50,
        # Private key is required for authenticated actions; use a dummy key for test (assumed accepted in test env)
        "private_key": "test_private_key_1234567890abcdef"
    }

    created_proposal_id = None

    try:
        # Step 1: Get existing proposals
        response = requests.get(proposals_url, headers=HEADERS, timeout=TIMEOUT)
        assert response.status_code == 200, f"Expected 200 OK from {proposals_url}, got {response.status_code}"
        proposals = response.json()
        assert isinstance(proposals, list), "/dao/proposals response is not a list"

        # If there are proposals already, verify structure of the first few items (if any)
        if proposals:
            for proposal in proposals[:5]:
                assert isinstance(proposal, dict), "Proposal item is not a dictionary"
                # Check presence and type of key fields
                assert "id" in proposal and isinstance(proposal["id"], str), "Proposal missing 'id' or it is not a string"
                assert "title" in proposal and isinstance(proposal["title"], str), "Proposal missing 'title' or it is not a string"
                assert "description" in proposal and isinstance(proposal["description"], str), "Proposal missing 'description' or it is not a string"
                assert "proposal_type" in proposal and isinstance(proposal["proposal_type"], str), "Proposal missing 'proposal_type' or it is not a string"
                assert "voting_type" in proposal and isinstance(proposal["voting_type"], str), "Proposal missing 'voting_type' or it is not a string"
                assert "duration" in proposal and isinstance(proposal["duration"], int), "Proposal missing 'duration' or it is not an int"
                assert "threshold" in proposal and isinstance(proposal["threshold"], int), "Proposal missing 'threshold' or it is not an int"
            return  # Test passed with existing proposals

        # Step 2: No proposals found, create one
        create_response = requests.post(proposal_create_url, json=new_proposal_payload, headers=HEADERS, timeout=TIMEOUT)
        assert create_response.status_code == 200, f"Expected 200 OK from {proposal_create_url}, got {create_response.status_code}"
        create_resp_json = create_response.json()
        assert isinstance(create_resp_json, dict), "Response from creating proposal is not a JSON object"
        created_proposal_id = create_resp_json.get("id")
        assert created_proposal_id and isinstance(created_proposal_id, str), "Created proposal response missing valid 'id'"

        # Sleep briefly to ensure proposal is indexed/available
        time.sleep(1)

        # Step 3: Get proposals again, verify newly created proposal is included
        response_after_create = requests.get(proposals_url, headers=HEADERS, timeout=TIMEOUT)
        assert response_after_create.status_code == 200, f"Expected 200 OK on second GET from {proposals_url}, got {response_after_create.status_code}"
        proposals_after_create = response_after_create.json()
        assert isinstance(proposals_after_create, list), "/dao/proposals response after create is not a list"

        ids = [p.get("id") for p in proposals_after_create if isinstance(p, dict)]
        assert created_proposal_id in ids, "Created proposal ID not found in list of proposals after creation"

        for proposal in proposals_after_create:
            if proposal.get("id") == created_proposal_id:
                # Verify fields match creation payload
                assert proposal.get("title") == new_proposal_payload["title"], "Proposal title mismatch"
                assert proposal.get("description") == new_proposal_payload["description"], "Proposal description mismatch"
                assert proposal.get("proposal_type") == new_proposal_payload["proposal_type"], "Proposal type mismatch"
                assert proposal.get("voting_type") == new_proposal_payload["voting_type"], "Voting type mismatch"
                assert isinstance(proposal.get("duration"), int), "Proposal duration is missing or not int"
                assert isinstance(proposal.get("threshold"), int), "Proposal threshold is missing or not int"
                break
        else:
            assert False, "Created proposal not found in proposals list after creation"

    finally:
        # Cleanup: delete the created proposal if possible
        # Note: No delete endpoint provided in PRD, so no delete step possible.
        # If deletion endpoint existed, we would attempt cleanup here.
        pass

test_get_all_governance_proposals()
