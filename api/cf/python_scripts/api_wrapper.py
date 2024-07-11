import os
import uuid
import shutil
import argparse
from cf import (
    main as cf_main,
    get_default_ua,
    DEFAULT_TARGET_URL,
    CfError,
    __version__,
)
from fastapi import (
    FastAPI,
    responses,
)
from pydantic import (
    BaseModel,
)

class CfArgs(BaseModel):
    attempts: int = 3
    headless: bool = True
    target_url: str = DEFAULT_TARGET_URL
    user_agent: str = get_default_ua()

app = FastAPI()

CHROME_PATH = shutil.which("google-chrome-stable")
if CHROME_PATH is None:
    raise FileNotFoundError("Google Chrome Stable not found")

def __read_log(log_path: str) -> str:
    if not os.path.exists(log_path):
        return ""

    try:
        with open(log_path, "r", encoding="utf-8") as f:
            return f.read()
    finally:
        try:
            os.remove(log_path)
        except:
            pass

@app.post("/", response_class=responses.JSONResponse)
def bypass(data: CfArgs):
    random_log_path = f"cf-{uuid.uuid4()}.log"
    try:
        values = cf_main(
            args=argparse.Namespace(
                attempts=data.attempts,
                virtual_display=True,
                test_connection=False,
                log_path=random_log_path,
                browser_path=CHROME_PATH,
                headless=data.headless,
                target_url=data.target_url,
                user_agent=data.user_agent,
            ),
        )
    except CfError:
        return responses.JSONResponse(
            content={"error": str(CfError)},
            status_code=500,
        )
    except Exception as e:
        return responses.JSONResponse(
            content={
                "error": "An unknown error occurred",
                "exception": str(e),
                "logs": __read_log(random_log_path),
            },
            status_code=500,
        )

    return {
        "log": __read_log(random_log_path),
        "values": values,
    }

@app.get("/", response_class=responses.JSONResponse)
def root():
    return {"version": __version__}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080, limit_concurrency=2)
