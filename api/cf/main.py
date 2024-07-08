# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass Cloudflare protection using DrissionPage.

Note: Logic based on https://github.com/sarperavci/CloudflareBypassForScraping
"""

import sys
import json
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
    return ChromiumPage(addr_or_opts=options)

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
        "-t", 
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

def validate_url(url: str) -> bool:
    if not url_validator(url):
        print(f"input error: invalid url, {url}, provided")
        sys.exit(1)

def validate_browser_path(browser_path_value: str) -> bool:
    try:
        browser_path = pathlib.Path(browser_path_value).resolve()
    except TypeError:
        print(f"input error: invalid browser path, {browser_path}, provided")
        sys.exit(1)

    if not browser_path.exists():
        print(f"input error: provided browser path, {browser_path}, does not exist")
        sys.exit(1)

    if not browser_path.is_file():
        print(f"input error: provided browser path, {browser_path}, is not a file")
        sys.exit(1)

def main(args_parser: argparse.ArgumentParser) -> None:
    args = args_parser.parse_args()
    browser_path = args.browser_path
    headless = args.headless
    target_url = args.target_url
    ua = args.user_agent

    validate_browser_path(browser_path)
    validate_url(target_url)

    driver = get_driver(browser_path, ua, headless)
    try:
        driver.get(target_url)
        cf_logic.bypass_cf(driver, target_url)
        cookies = driver.cookies(as_dict=True, all_domains=True)
    finally:
        driver.quit()

    with tempfile.NamedTemporaryFile(mode="w", delete=False, delete_on_close=False) as f:
        json.dump(cookies, f)
        print(f"cookies saved to {f.name}")

if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(levelname)s - %(message)s",
    )
    main(create_arg_parser())
