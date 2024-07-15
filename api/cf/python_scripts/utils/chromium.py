"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import os
import atexit
import shutil
import tempfile

import _types
import _logger
import constants
from .general import (
    get_base_url,
)

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

def get_default_ua() -> str:
    match constants.PLATFORM_NAME:
        case "Linux":
            ua_os = "X11; Linux x86_64"
        case "Darwin":
            ua_os = "Macintosh; Intel Mac OS X 10_15_7"
        case _:
            ua_os = "Windows NT 10.0; Win64; x64"
    return f"Mozilla/5.0 ({ua_os}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

def __stop_listener(page: ChromiumPage) -> None:
    _logger.get_logger().info("Stopping listener...")
    page.listen.stop()

def start_listener(page: ChromiumPage, target_url: str) -> None:
    """
    Start a listener on the specified target URL.

    Mainly used to listen for GET responses that is returning a HTML document from the target URL. 

    Note that `.listen.stop` will be registered with atexit to ensure the listener is stopped so you don't have to manually call `.listen.stop()`.

    Args:
        page (ChromiumPage): 
            ChromiumPage object.
        target_url (str): 
            Target URL to listen on.

    Returns:
        None
    """
    page.listen.start(
        targets=get_base_url(target_url), 
        method="GET", 
        res_type="Document",
    )
    atexit.register(__stop_listener, page=page)

def save_cookies(cookies: _types.Cookies) -> None:
    logger = _logger.get_logger()
    logger.info("Saving cookies...")
    with tempfile.NamedTemporaryFile(mode="w", prefix="kjhjason-cf-", delete=False, delete_on_close=False) as f:
        serialised_cookies = orjson.dumps(
            cookies, 
            option=orjson.OPT_NON_STR_KEYS,
        )
        f.write(serialised_cookies.decode("utf-8"))

        msg = f"cookies saved to {f.name}"
        print(msg)
        logger.info(msg)

@atexit.register
def __remove_navigator_js() -> None:
    if os.path.exists(constants.NAVIGATOR_JS_PATH):
        os.remove(constants.NAVIGATOR_JS_PATH)

def __create_navigator_js_with_os_name(os_name: str) -> None:
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

def __close_chromium_page(page: ChromiumPage) -> None:
    _logger.get_logger().info("Closing browser...")
    page.quit()

def get_chromium_page(
    browser_path: str, 
    os_name: str, 
    user_agent: str, 
    headless: bool, 
    no_sandbox: bool = False,
) -> ChromiumPage:
    """
    Get an initialised ChromiumPage object with the specified options.

    Note that `.quit()` will be registered with atexit to ensure the browser is closed so you don't have to manually call `.quit()`.

    Args:
        browser_path (str): 
            Path to the browser executable.
        os_name (str): 
            OS name to spoof in navigator.platform (should match the user-agent).
        user_agent (str): 
            User-agent to use.
        headless (bool): 
            Run the browser in headless mode.
        no_sandbox (bool, optional): 
            Run the browser with no-sandbox mode. Defaults to False.
            Use this if running as a non-root user on unix-like systems.

    Returns:
        ChromiumPage: ChromiumPage object.

    Raises:
        drission_errors.BrowserConnectError: 
            If unable to connect to the browser.
    """
    options = ChromiumOptions()
    options.auto_port()
    options.set_paths(browser_path=browser_path)
    options.headless(headless)
    options.set_user_agent(user_agent)

    if not no_sandbox and constants.IS_UNIX and (os.getenv("KJHJASON_CF_SANDBOX") == "1" or os.geteuid() != 0):
        # --no-sandbox is required if not running as root user.
        # Otherwise, the browser may have errors trying to launch as root.
        no_sandbox = True

    logger = _logger.get_logger()
    if no_sandbox:
        logger.info("Running with no-sandbox mode...")
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
        "--disable-infobars",
        "--disable-suggestions-ui",
        "--hide-crash-restore-bubble",
        f"--window-size={constants.WINDOW_SIZE_X},{constants.WINDOW_SIZE_Y}",
        "--accept-lang=en-US",
    )
    for arg in args:
        options.set_argument(arg)

    if os.getenv("KJHJASON_CF_ADD_NAV_EXT") == "1" or os_name != constants.PLATFORM_NAME.lower():
        __create_navigator_js_with_os_name(os_name)
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

    atexit.register(__close_chromium_page, page=page)
    return page
