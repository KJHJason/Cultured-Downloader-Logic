"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import os
import shutil
import typing
import logging
import tempfile
import subprocess

import _types
import errors
import constants

import orjson
from DrissionPage import (
    ChromiumPage, 
    ChromiumOptions,
    errors as drission_errors,
)

def get_default_chrome_path() -> str:
    match constants.PLATFORM_NAME:
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

def check_container(app_key: str, logger: logging.Logger) -> None | typing.NoReturn:
    # Mainly just for obfuscation purposes to make it harder to run the script in a container.
    if constants.IS_DOCKER and app_key != "fzN9Hvkb9s+mwPGCDd5YFnLiqKx8WhZfWoZE5nZC":
        errors.handle_err("Failed to connect to browser...", logger)
        return

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
    match constants.PLATFORM_NAME:
        case "Linux":
            ua_os = "X11; Linux x86_64"
        case "Darwin":
            ua_os = "Macintosh; Intel Mac OS X 10_15_7"
        case _:
            ua_os = "Windows NT 10.0; Win64; x64"
    return f"Mozilla/5.0 ({ua_os}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

def edit_navigator_js_with_os_name(os_name: str) -> None:
    if os.path.exists(constants.NAVIGATOR_JS_PATH):
        return

    match os_name:
        case "linux":
            nav_platform = "Linux x86_64"
        case "darwin":
            nav_platform = "MacIntel"
        case "windows":
            nav_platform = "Win32"
        case _:
            raise ValueError("Unsupported OS Name")

    with open(constants.NAVIGATOR_EXT_DIR + "/base.js", "r", encoding="utf-8") as f:
        navigator_js = f.read()

    navigator_js = navigator_js.replace("<OS_NAME>", nav_platform, 1)
    with open(constants.NAVIGATOR_JS_PATH, "w", encoding="utf-8") as f:
        f.write(navigator_js)

def get_chromium_page(browser_path: str, os_name: str, user_agent: str, headless: bool, no_sandbox: bool = False) -> ChromiumPage:
    options = ChromiumOptions()
    options.auto_port()
    options.set_paths(browser_path=browser_path)
    options.headless(headless)
    options.set_user_agent(user_agent)

    if not no_sandbox and constants.IS_UNIX and (os.environ.get("KJHJASON_CF_SANDBOX") == "1" or os.geteuid() != 0):
        # --no-sandbox is required if not running as root user.
        # Otherwise, the browser may have errors trying to launch as root.
        no_sandbox = True

    if no_sandbox:
        logging.info("Running with no-sandbox mode...")
        options.set_argument("--no-sandbox")

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

    if os.environ.get("KJHJASON_CF_ADD_NAV_EXT") == "1" or os_name != constants.PLATFORM_NAME.lower():
        edit_navigator_js_with_os_name(os_name)
        options.add_extension(constants.NAVIGATOR_EXT_DIR)

    try:
        page = ChromiumPage(addr_or_opts=options)
    except drission_errors.BrowserConnectError as e:
        if constants.IS_UNIX and not no_sandbox:
            # Try again with --no-sandbox flag
            return get_chromium_page(
                browser_path=browser_path, 
                os_name=os_name,
                user_agent=user_agent, 
                headless=headless, 
                no_sandbox=True,
            )
        raise e

    if headless or os.environ.get("KJHJASON_CF_SET_MAX_WINDOW") == "1":
        page.set.window.max()
    return page
