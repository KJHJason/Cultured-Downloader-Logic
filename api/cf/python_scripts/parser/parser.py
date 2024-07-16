"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import typing
import pathlib
import argparse
import functools

import utils
import constants

import validators.url as url_validator

def parse_bool(s: str) -> bool:
    return s == "1" or s == "true" or s == "True"

@functools.cache
def create_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="CF Bypass", 
        description="CF Bypass Script by KJHJason",
    )
    parser.add_argument(
        "-v",
        "--version", 
        action="version", 
        version=f"%(prog)s v{constants.__version__}",
    )
    parser.add_argument(
        "-os",
        "--os-name",
        type=str,
        choices=constants.OS_CHOICES,
        default=constants.PLATFORM_NAME.lower(),
        help="Define the OS for navigator.platform (Required if spoofing user-agent of another OS)",
    )
    parser.add_argument(
        "-ak",
        "--app-key",
        type=str,
        default="",
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
        default=f"cf-{constants.__version__}.log",
    )
    parser.add_argument(
        "--cookie-path",
        type=str,
        help="Path to save the cookies JSON file (Defaults to your temp directory)",
        default=""
    )
    parser.add_argument(
        "--browser-path", 
        type=str, 
        help="Path to the Google Chrome browser executable", 
        default=utils.get_default_chrome_path(),
    )
    parser.add_argument(
        "--headless", 
        type=str,
        choices=constants.ARGS_BOOLEAN_CHOICE,
        help="Run the browser in headless mode",
        default=str(constants.IS_DOCKER),
    )
    parser.add_argument(
        "--target-url", 
        type=str, 
        help="URL to visit and bypass", 
        default=constants.DEFAULT_TARGET_URL,
    )
    parser.add_argument(
        "-ua", 
        "--user-agent", 
        type=str,
        help="User-Agent to use", 
        default=utils.get_default_ua(),
    )
    return parser

def validate_headless(headless: bool, virtual_display: bool) -> None | typing.NoReturn:
    if constants.IS_DOCKER and not headless and not virtual_display:
        utils.handle_err("input error: headless mode cannot be disabled without virtual displays in docker, use --virtual-display or set --headless=1 instead")

def validate_url(url: str) -> None | typing.NoReturn:
    if not url_validator(url):
        utils.handle_err(f"input error: invalid url, {url}, provided")

def validate_browser_path(browser_path_value: str) -> None | typing.NoReturn:
    try:
        browser_path = pathlib.Path(browser_path_value).resolve()
    except TypeError:
        utils.handle_err(f"input error: invalid browser path, {browser_path}, provided")

    if not browser_path.exists():
        utils.handle_err(f"input error: provided browser path, {browser_path}, does not exist")

    if not browser_path.is_file():
        utils.handle_err(f"input error: provided browser path, {browser_path}, is not a file")
