import os
import shutil
import logging
import tempfile
import platform
import subprocess

import _types

import orjson
from DrissionPage import (
    ChromiumPage, 
    ChromiumOptions,
    errors as drission_errors,
)

def get_default_chrome_path() -> str:
    match platform.system():
        case "Linux":
            return shutil.which("google-chrome") or "/usr/bin/google-chrome"
        case "Darwin":
            return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
        case "Windows":
            return r"C:/Program Files/Google/Chrome/Application/chrome.exe"
        case _:
            raise ValueError("Unsupported OS")

def check_for_xvfb(logger: logging.Logger) -> bool:
    if shutil.which("xvfb-run") is not None:
        return True

    try:
        subprocess.run(["Xvfb", "-help"], check=True)
    except subprocess.CalledProcessError:
        logger.warning("xvfb-run not found, ignoring --virtual-display flag...")
        return False
    return True

def save_cookies(cookies: _types.Cookies, logger: logging.Logger) -> None:
    logger.info("Saving cookies...")
    with tempfile.NamedTemporaryFile(mode="w", prefix="kjhjason-cf-", delete=False, delete_on_close=False) as f:
        serialised_cookies = orjson.dumps(
            cookies, 
            option=orjson.OPT_NON_STR_KEYS | orjson.OPT_SERIALIZE_NUMPY,
        )
        f.write(serialised_cookies.decode("utf-8"))

        msg = f"cookies saved to {f.name}"
        print(msg)
        logger.info(msg)

def get_default_ua() -> str:
    match platform.system():
        case "Linux":
            ua_os = "X11; Linux x86_64"
        case "Darwin":
            ua_os = "Macintosh; Intel Mac OS X 10_15_7"
        case _:
            ua_os = "Windows NT 10.0; Win64; x64"
    return f"Mozilla/5.0 ({ua_os}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

def get_chromium_page(browser_path: str, ua: str, headless: bool, no_sandbox: bool = False) -> ChromiumPage:
    options = ChromiumOptions()
    options.auto_port()
    options.set_paths(browser_path=browser_path)
    options.headless(headless)
    options.set_user_agent(ua)

    os_name = platform.system()
    is_unix = os_name == "Linux" or os_name == "Darwin"
    if is_unix and (os.environ.get("KJHJASON_CF_SANDBOX") == "1" or os.geteuid() != 0):
        # --no-sandbox is required if not running as root user.
        # Otherwise, the browser may have errors trying to launch as root.
        options.set_argument("--no-sandbox")
        logging.info("Running with no-sandbox mode...")
        no_sandbox = True

    args = (
        "--no-first-run",
        "--force-color-profile=srgb",
        "--metrics-recording-only",
        "--password-store=basic",
        "--use-mock-keychain",
        "--export-tagged-pdf",
        "--no-default-browser-check",
        "--disable-background-mode",
        "--enable-features=NetworkService,NetworkServiceInProcess,LoadCryptoTokenExtension,PermuteTLSExtensions",
        "--disable-features=FlashDeprecationWarning,EnablePasswordsAccountStorage",
        "--deny-permission-prompts",
        "--disable-gpu",
        "--accept-lang=en-US",
    )
    for arg in args:
        options.set_argument(arg)

    if os.environ.get("KJHJASON_CF_ADD_NAV_EXT") == "1":
        options.add_extension("./extensions/Navigator/")

    try:
        page = ChromiumPage(addr_or_opts=options)
    except drission_errors.BrowserConnectError as e:
        if is_unix and not no_sandbox:
            page.quit()
            # Try again with --no-sandbox flag
            return get_chromium_page(browser_path, ua, headless, no_sandbox=True)
        raise e

    if headless or os.environ.get("KJHJASON_CF_SET_MAX_WINDOW") == "1":
        page.set.window.max()
    return page
