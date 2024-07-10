# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass Cloudflare protection using DrissionPage."""

import sys
import json
import typing
import tempfile
import logging
import pathlib
import platform
import argparse
import cf_logic
from DrissionPage import (
    ChromiumPage, 
    ChromiumOptions,
)
import validators.url as url_validator

__version__ = "0.1.0"

def get_driver(browser_path: str, ua: str, headless: bool) -> ChromiumPage:
    options = ChromiumOptions()
    options.set_paths(browser_path=browser_path)
    options.headless(headless)
    options.set_user_agent(ua)

    args = (
        "-no-first-run",
        "-force-color-profile=srgb",
        "-metrics-recording-only",
        "-password-store=basic",
        "-use-mock-keychain",
        "-export-tagged-pdf",
        "-no-default-browser-check",
        "-disable-background-mode",
        "-enable-features=NetworkService,NetworkServiceInProcess,LoadCryptoTokenExtension,PermuteTLSExtensions",
        "-disable-features=FlashDeprecationWarning,EnablePasswordsAccountStorage",
        "-deny-permission-prompts",
        "-disable-gpu",
    )
    for arg in args:
        options.set_argument(arg)

    driver = ChromiumPage(addr_or_opts=options)
    if headless:
        driver.set.window.max()
    return driver

def get_default_ua() -> str:
    ua_os = ""
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
        prog="Cloudflare Bypass", 
        description="Cloudflare Bypass Script by KJHJason",
    )
    parser.add_argument(
        "-v",
        "--version", 
        action="version", 
        version=f"%(prog)s v{__version__}",
    )
    parser.add_argument(
        "--attempts",
        type=int,
        help="Number of attempts to try and bypass Cloudflare (0 for infinite attempts)",
        default=0,
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
        default="https://nopecha.com/demo/cloudflare",
    )
    parser.add_argument(
        "-ua", 
        "--user-agent", 
        type=str,
        help="User-Agent to use", 
        default=get_default_ua(),
    )
    return parser

def __handle_err(msg: str) -> typing.NoReturn:
    print(msg)
    logging.error(msg)
    sys.exit(1)

def validate_url(url: str) -> bool:
    if not url_validator(url):
        __handle_err(f"input error: invalid url, {url}, provided")

def validate_browser_path(browser_path_value: str) -> bool:
    try:
        browser_path = pathlib.Path(browser_path_value).resolve()
    except TypeError:
        __handle_err(f"input error: invalid browser path, {browser_path}, provided")

    if not browser_path.exists():
        __handle_err(f"input error: provided browser path, {browser_path}, does not exist")

    if not browser_path.is_file():
        __handle_err(f"input error: provided browser path, {browser_path}, is not a file")

def save_cookies(cookies: dict) -> None:
    logging.info("Saving cookies...")
    with tempfile.NamedTemporaryFile(mode="w", prefix="kjhjason-cf-", delete=False, delete_on_close=False) as f:
        json.dump(cookies, f)
        msg = f"cookies saved to {f.name}"
        print(msg)
        logging.info(msg)

def main(args_parser: argparse.ArgumentParser) -> None:
    args = args_parser.parse_args()

    log_path_arg: str = args.log_path
    log_path = pathlib.Path(log_path_arg).resolve()
    if not (log_path_dir := log_path.parent).exists():
        log_path_dir.mkdir(parents=True)

    logging.basicConfig(
        filename=log_path,
        level=logging.INFO,
        format="%(asctime)s - %(levelname)s - %(message)s",
    )

    attempts: int = args.attempts
    browser_path: str = args.browser_path
    headless: bool = args.headless
    target_url: str = args.target_url
    ua: str = args.user_agent

    validate_browser_path(browser_path)
    validate_url(target_url)

    logging.info("Starting Cloudflare Bypass...")
    driver = get_driver(browser_path, ua, headless)
    try:
        driver.get(target_url)
        if cf_logic.bypass_cf(driver, target_url, attempts):
            cookies = driver.cookies(as_dict=False, all_domains=False, all_info=True)
            save_cookies(cookies)
        else:
            logging.error("Failed to bypass Cloudflare protection, max attempts reached...")
    except KeyboardInterrupt:
        logging.info("Script interrupted.")
    except Exception as e:
        logging.error(f"An error occurred:\n{e}\n")
        raise e
    finally:
        logging.info("Closing browser...")
        driver.quit()

if __name__ == "__main__":
    main(create_arg_parser())
