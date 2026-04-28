import requests
import json
import sseclient
import threading
import time

def test_mcp():
    url = "http://localhost:8000/sse"
    print(f"Connecting to {url}...")
    
    session_id = None
    connected_event = threading.Event()

    def listen_sse():
        nonlocal session_id
        try:
            response = requests.get(url, stream=True)
            client = sseclient.SSEClient(response)
            for event in client.events():
                if event.event == 'endpoint':
                    endpoint_url = event.data
                    session_id = endpoint_url.split("sessionId=")[1]
                    print(f"Found Session ID: {session_id}")
                    connected_event.set()
                # Keep the connection open
        except Exception as e:
            print(f"SSE Connection error: {e}")

    sse_thread = threading.Thread(target=listen_sse, daemon=True)
    sse_thread.start()
    
    # Wait for session_id
    if not connected_event.wait(timeout=10):
        print("Timeout waiting for session ID")
        return

    # Now send the call_tool request
    message_url = f"http://localhost:8000/message?sessionId={session_id}"
    payload = {
        "jsonrpc": "2.0",
        "id": "1",
        "method": "tools/call",
        "params": {
            "name": "kairo_create_task",
            "arguments": {
                "title": "Task via Python MCP Test",
                "description": "This task was added by a Python script simulating an MCP client connecting to the SSE server.",
                "status": "todo",
                "priority": 1,
                "tags": "test, python, mcp"
            }
        }
    }
    
    print(f"Sending request to {message_url}...")
    post_resp = requests.post(message_url, json=payload)
    print(f"POST Status: {post_resp.status_code}")
    print(f"POST Response: {post_resp.text}")

    # Give it a second to process and maybe receive a response back on SSE if we wanted
    time.sleep(1)

if __name__ == "__main__":
    test_mcp()
