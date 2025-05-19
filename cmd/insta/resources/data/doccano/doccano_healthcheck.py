import http.client
import sys
import os

# Doccano runs on port 8000 by default within its container
HOST = os.getenv('DOCCANO_INTERNAL_HOST', 'localhost') # localhost from within the container
PORT = int(os.getenv('DOCCANO_INTERNAL_PORT', '8000'))
PATH = "/v1/health"

print(f"Healthcheck: Attempting to connect to http://{HOST}:{PORT}{PATH}")

try:
    conn = http.client.HTTPConnection(HOST, PORT, timeout=3) # Short timeout
    conn.request("GET", PATH)
    response = conn.getresponse()
    print(f"Healthcheck: Received status {response.status}")
    if response.status == 200:
        # Optionally, read response.read() if specific content is expected
        print("Healthcheck: Success (status 200)")
        conn.close()
        sys.exit(0) # Success
    else:
        print(f"Healthcheck: Failed (status {response.status})")
        conn.close()
        sys.exit(1) # Failure
except ConnectionRefusedError:
    print("Healthcheck: Failed (Connection refused)")
    sys.exit(1)
except http.client.HTTPException as e:
    print(f"Healthcheck: Failed (HTTPException: {e})")
    sys.exit(1)
except TimeoutError: # For socket timeout
    print("Healthcheck: Failed (Timeout)")
    sys.exit(1)
except Exception as e:
    print(f"Healthcheck: Failed (Unknown error: {e})")
    sys.exit(1) 