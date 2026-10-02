import time
import httpx
from app.config import settings
from app.queue.redis_worker import start_worker

def wait_for_go_service(url: str, timeout: int = 10):
    start = time.time()
    while time.time() - start < timeout:
        try:
            resp = httpx.get(f"{url}/health", timeout=3.0)
            if resp.status_code == 200:
                print(f"Go service ready at {url}")
                return
        except Exception:
            pass
        print("Waiting for Go service...")
        time.sleep(2)
    print("Go service wait timeout or skipped.")

if __name__ == "__main__":
    print("Starting Python Agent Worker...")
    wait_for_go_service(settings.GO_SERVICE_URL)
    start_worker()
