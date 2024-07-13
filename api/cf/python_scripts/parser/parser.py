import typing
import logging
import pathlib
import argparse

import utils
import errors
import constants

import validators.url as url_validator

def parse_bool(s: str) -> bool:
    return s == "true" or s == "True" or s == "1"

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
        "--browser-path", 
        type=str, 
        help="Path to the Google Chrome browser executable", 
        default=utils.get_default_chrome_path(),
    )
    parser.add_argument(
        "--headless", 
        type=parse_bool,
        choices=["true", "True", "1", "false", "False", "0"],
        help="Run the browser in headless mode",
        default=constants.IS_DOCKER,
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

def validate_headless(headless:bool, logger: logging.Logger) -> None | typing.NoReturn:
    if constants.IS_DOCKER and not headless:
        errors.handle_err("input error: headless mode cannot be used in docker, use --virtual-display or set --headless=false instead", logger)

def validate_url(url: str, logger: logging.Logger) -> None | typing.NoReturn:
    if not url_validator(url):
        errors.handle_err(f"input error: invalid url, {url}, provided", logger)

def validate_browser_path(browser_path_value: str, logger: logging.Logger) -> None | typing.NoReturn:
    try:
        browser_path = pathlib.Path(browser_path_value).resolve()
    except TypeError:
        errors.handle_err(f"input error: invalid browser path, {browser_path}, provided", logger)

    if not browser_path.exists():
        errors.handle_err(f"input error: provided browser path, {browser_path}, does not exist", logger)

    if not browser_path.is_file():
        errors.handle_err(f"input error: provided browser path, {browser_path}, is not a file", logger)
