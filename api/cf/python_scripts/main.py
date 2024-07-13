# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass CF protection using DrissionPage."""

import sys
import logging
import pathlib
import platform
import argparse

import test
import logic
import utils
import _types
import errors
import parser

from DrissionPage import (
    errors as drission_errors,
)

def __main(browser_path: str, ua: str, headless: bool, target_url: str, attempts: int, test_connection: bool, logger: logging.Logger):
    logger.info("Starting CF Bypass...")

    try:
        page = utils.get_chromium_page(browser_path, ua, headless)
    except drission_errors.BrowserConnectError as e:
        logger.error(f"Failed to connect to browser:\n{e}\n")
        if test_connection:
            raise test.Results(success=False)
        raise e
    else:
        if test_connection:
            page.quit()
            raise test.Results(success=True)

    try:
        page.listen.start(
            targets=logic.get_base_url(target_url), 
            method="GET", 
            res_type="Document",
        )
        page.get(target_url)
        if logic.bypass_cf(page, attempts, logger):
            cookies: _types.Cookies = page.cookies(as_dict=False, all_domains=False, all_info=True)
            utils.save_cookies(cookies, logger)
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

def main(args: argparse.Namespace) -> _types.Cookies:
    log_path_arg: str = args.log_path
    log_path = pathlib.Path(log_path_arg).resolve()
    if not (log_path_dir := log_path.parent).exists():
        log_path_dir.mkdir(parents=True)

    logic.configure_logger(log_path)
    logger = logic.get_logger()

    os_name = platform.system()
    virtual_display: bool = args.virtual_display
    if os_name not in ("Linux", "Darwin",) and virtual_display:
        logger.warning("Virtual display is only supported on unix-like systems, ignoring --virtual-display flag...")
        virtual_display = False
    elif virtual_display and not utils.check_for_xvfb(logger):
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

    parser.validate_headless(headless, logger)
    parser.validate_url(target_url, logger)
    parser.validate_browser_path(browser_path, logger)

    if not virtual_display:
        try:
            __main(
                browser_path=browser_path,
                ua=ua,
                headless=headless,
                target_url=target_url,
                attempts=attempts,
                logger=logger,
                test_connection=test_connection,
            )
        except test.Results as e:
            e.handle_result(logger)
        return

    import pyvirtualdisplay
    try:
        with pyvirtualdisplay.Display(visible=0, backend="xvfb", size=(1024, 768)):
            __main(
                browser_path=browser_path,
                ua=ua,
                headless=False,
                target_url=target_url,
                attempts=attempts,
                logger=logger,
                test_connection=test_connection,
            )
    except test.Results as e:
        e.handle_result(logger)

if __name__ == "__main__":
    try:
        main(args=parser.create_arg_parser().parse_args())
    except errors.CfError:
        sys.exit(1)
