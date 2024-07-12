# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass CF protection using DrissionPage."""

import os
import sys
import json
import shutil
import typing
import tempfile
import logging
import pathlib
import platform
import argparse
import subprocess
import cf_logic
from DrissionPage import (
    ChromiumPage, 
    ChromiumOptions,
    errors as drission_errors,
)
import validators.url as url_validator

__version__ = "0.1.0"
DEFAULT_TARGET_URL = "https://nopecha.com/demo/cloudflare"

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

    try:
        page = ChromiumPage(addr_or_opts=options)
    except drission_errors.BrowserConnectError as e:
        if is_unix and not no_sandbox:
            # Try again with --no-sandbox flag
            return get_chromium_page(browser_path, ua, headless, no_sandbox=True)
        raise e

    if headless:
        page.set.window.max()
    return page

def get_default_ua() -> str:
    match platform.system():
        case "Linux":
            ua_os = "X11; Linux x86_64"
        case "Darwin":
            ua_os = "Macintosh; Intel Mac OS X 10_15_7"
        case _:
            ua_os = "Windows NT 10.0; Win64; x64"
    return f"Mozilla/5.0 ({ua_os}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

def create_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="CF Bypass", 
        description="CF Bypass Script by KJHJason",
    )
    parser.add_argument(
        "-v",
        "--version", 
        action="version", 
        version=f"%(prog)s v{__version__}",
    )
    parser.add_argument(
        "-tc",
        "--test-connection",
        action="store_true",
        help="Run the script to test if it works as unix system commonly faces BrowserConnectError",
        default=False,
    )
    parser.add_argument(
        "--attempts",
        type=int,
        help="Number of attempts to try and bypass CF (0 for infinite attempts)",
        default=0,
    )
    parser.add_argument(
        "--virtual-display", 
        action="store_true", 
        help="Run the browser in a virtual display using xvfb (Linux only)",
        default=False,
    )
    parser.add_argument(
        "--log-path",
        type=str,
        help="Path to save log file",
        default=f"cf-{__version__}.log",
    )
    parser.add_argument(
        "--browser-path", 
        type=str, 
        help="Path to the Google Chrome browser executable", 
        default=r"C:/Program Files/Google/Chrome/Application/chrome.exe",
    )
    parser.add_argument(
        "--headless", 
        action="store_true", 
        help="Run the browser in headless mode",
        default=False,
    )
    parser.add_argument(
        "--target-url", 
        type=str, 
        help="URL to visit and bypass", 
        default=DEFAULT_TARGET_URL,
    )
    parser.add_argument(
        "-ua", 
        "--user-agent", 
        type=str,
        help="User-Agent to use", 
        default=get_default_ua(),
    )
    return parser

class CfError(Exception):
    def __init__(self, msg: str) -> None:
        self.msg = msg

    def __str__(self) -> str:
        return self.msg

class TestResults(Exception):
    def __init__(self, success: bool) -> None:
        self.success = success

    def __str__(self) -> str:
        return f"Test {'succeeded' if self.success else 'failed'}"

    def handle_result(self, logger: logging.Logger) -> typing.NoReturn:
        if self.success:
            logger.info("Test succeeded")
            print("Test succeeded")
            sys.exit(0)
        else:
            logger.error("Test failed")
            print("Test failed")
            sys.exit(1)

def __handle_err(msg: str, logger: logging.Logger) -> None:
    print(msg)
    logger.error(msg)
    raise CfError(msg)

def validate_url(url: str, logger: logging.Logger) -> bool:
    if not url_validator(url):
        __handle_err(f"input error: invalid url, {url}, provided", logger)

def validate_browser_path(browser_path_value: str, logger: logging.Logger) -> bool:
    try:
        browser_path = pathlib.Path(browser_path_value).resolve()
    except TypeError:
        __handle_err(f"input error: invalid browser path, {browser_path}, provided", logger)

    if not browser_path.exists():
        __handle_err(f"input error: provided browser path, {browser_path}, does not exist", logger)

    if not browser_path.is_file():
        __handle_err(f"input error: provided browser path, {browser_path}, is not a file", logger)

def save_cookies(cookies: list[dict[str, str | float | bool | int]], logger: logging.Logger) -> None:
    logger.info("Saving cookies...")
    with tempfile.NamedTemporaryFile(mode="w", prefix="kjhjason-cf-", delete=False, delete_on_close=False) as f:
        json.dump(cookies, f)
        msg = f"cookies saved to {f.name}"
        print(msg)
        logger.info(msg)

def __main(browser_path: str, ua: str, headless: bool, target_url: str, attempts: int, test_connection: bool, logger: logging.Logger) -> list[dict[str, str | float | bool | int]]:
    logger.info("Starting CF Bypass...")

    try:
        page = get_chromium_page(browser_path, ua, headless)
    except drission_errors.BrowserConnectError as e:
        logger.error(f"Failed to connect to browser:\n{e}\n")
        if test_connection:
            raise TestResults(success=False)
        raise e
    finally:
        if test_connection:
            page.quit()
            raise TestResults(success=True)

    cookies: list[dict[str, str | float | bool | int]] = []
    try:
        page.listen.start(
            targets=cf_logic.get_base_url(target_url), 
            method="GET", 
            res_type="Document",
        )
        page.get(target_url)
        if cf_logic.bypass_cf(page, attempts, logger):
            cookies = page.cookies(as_dict=False, all_domains=False, all_info=True)
            save_cookies(cookies, logger)
        else:
            logger.error("Failed to bypass CF protection, max attempts reached...")
    except KeyboardInterrupt:
        logger.info("Script interrupted.")
    except Exception as e:
        logger.error(f"An error occurred:\n{e}\n")
        raise e
    finally:
        logger.info("Closing browser...")
        page.listen.stop()
        page.quit()

    return cookies

def check_for_xvfb(logger: logging.Logger) -> bool:
    if shutil.which("xvfb-run") is not None:
        return True

    try:
        subprocess.run(["Xvfb", "-help"], check=True)
    except subprocess.CalledProcessError:
        logger.warning("xvfb-run not found, ignoring --virtual-display flag...")
        return False
    return True

def main(args: argparse.Namespace) -> list[dict[str, str | float | bool | int]]:
    log_path_arg: str = args.log_path
    log_path = pathlib.Path(log_path_arg).resolve()
    if not (log_path_dir := log_path.parent).exists():
        log_path_dir.mkdir(parents=True)

    cf_logic.configure_logger(log_path)
    logger = cf_logic.get_logger()

    os_name = platform.system()
    virtual_display: bool = args.virtual_display
    if os_name not in ("Linux", "Darwin",) and virtual_display:
        logger.warning("Virtual display is only supported on unix-like systems, ignoring --virtual-display flag...")
        virtual_display = False
    elif virtual_display and not check_for_xvfb(logger):
        virtual_display = False

    headless: bool = args.headless
    if headless and virtual_display:
        logger.warning("no need to use virtual display with headless mode, ignoring --virtual-display flag...")
        virtual_display = False

    test_connection: bool = args.test_connection
    attempts: int = args.attempts
    browser_path: str = args.browser_path
    target_url: str = args.target_url
    ua: str = args.user_agent

    validate_browser_path(browser_path, logger)
    validate_url(target_url, logger)

    if not virtual_display:
        try:
            return __main(
                browser_path=browser_path,
                ua=ua,
                headless=headless,
                target_url=target_url,
                attempts=attempts,
                logger=logger,
                test_connection=test_connection,
            )
        except TestResults as e:
            e.handle_result(logger)

    import pyvirtualdisplay
    try:
        with pyvirtualdisplay.Display(visible=0, backend="xvfb", size=(1024, 768)):
            return __main(
                browser_path=browser_path,
                ua=ua,
                headless=False,
                target_url=target_url,
                attempts=attempts,
                logger=logger,
                test_connection=test_connection,
            )
    except TestResults as e:
        e.handle_result(logger)

if __name__ == "__main__":
    try:
        main(args=create_arg_parser().parse_args())
    except CfError:
        sys.exit(1)
