import os

__version__ = "0.1.0"
DEFAULT_TARGET_URL = "https://nopecha.com/demo/cloudflare"
CF_WRAPPER_XPATH = ".cf-turnstile-wrapper"
LOGGER_NAME = "cf_bypass"
IS_DOCKER = os.environ.get("KJHJASON_CF_DOCKER") == "1"
