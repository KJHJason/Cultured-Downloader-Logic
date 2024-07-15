"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3

Simple script to bypass CF protection using DrissionPage.
"""

import pathlib
import argparse

import test
import logic
import utils
import _types
import parser
import _logger
import constants

from DrissionPage import (
    errors as drission_errors,
)

def __main(
    browser_path: str, 
    os_name: str,
    user_agent: str, 
    headless: bool, 
    target_url: str, 
    attempts: int, 
    test_connection: bool, 
    app_key: str = "",
) -> None:
    logger = _logger.get_logger()
    logger.info("Starting CF Bypass...")

    utils.check_container(app_key)
    try:
        page = utils.get_chromium_page(
            os_name=os_name,
            user_agent=user_agent, 
            headless=headless,
            browser_path=browser_path, 
        )
    except drission_errors.BrowserConnectError as e:
        logger.error(f"Failed to connect to browser:\n{e}\n")
        if test_connection:
            raise test.Results(success=False)
        raise e
    else:
        if test_connection:
            raise test.Results(success=True)

    try:
        utils.start_listener(page, target_url)
        page.get(target_url)
        if logic.bypass_cf(page, attempts, target_url):
            cookies: _types.Cookies = page.cookies(as_dict=False, all_domains=False, all_info=True)
            utils.save_cookies(cookies)
        else:
            logger.error("Failed to bypass CF protection, max attempts reached...")
    except KeyboardInterrupt:
        logger.info("Script interrupted.")
    except Exception as e:
        logger.error(f"An error occurred:\n{e}\n")
        raise e

def main(args: argparse.Namespace = parser.create_arg_parser().parse_args()) -> None:
    log_path_arg: str = args.log_path
    log_path = pathlib.Path(log_path_arg).resolve()
    if not (log_path_dir := log_path.parent).exists():
        log_path_dir.mkdir(parents=True)

    _logger.configure_logger(log_path)
    logger = _logger.get_logger()

    virtual_display: bool = args.virtual_display
    if not constants.IS_UNIX and virtual_display:
        logger.warning("Virtual display is only supported on unix-like systems, ignoring --virtual-display flag...")
        virtual_display = False
    elif virtual_display and not utils.check_for_xvfb():
        virtual_display = False

    headless_val: str = args.headless
    headless: bool = parser.parse_bool(headless_val)
    if headless and virtual_display:
        logger.warning("no need to use virtual display with headless mode, ignoring --virtual-display flag...")
        virtual_display = False

    os_name: str = args.os_name
    test_connection: bool = args.test_connection
    attempts: int = args.attempts
    browser_path: str = args.browser_path
    target_url: str = args.target_url
    user_agent: str = args.user_agent
    app_key: str = args.app_key

    parser.validate_headless(headless)
    parser.validate_url(target_url)
    parser.validate_browser_path(browser_path)

    if not virtual_display:
        try:
            __main(
                browser_path=browser_path,
                os_name=os_name,
                user_agent=user_agent,
                headless=headless,
                target_url=target_url,
                attempts=attempts,
                test_connection=test_connection,
                app_key=app_key,
            )
        except test.Results as e:
            e.handle_result()
        return

    import pyvirtualdisplay
    try:
        with pyvirtualdisplay.Display(visible=0, backend="xvfb", size=(constants.WINDOW_SIZE_X, constants.WINDOW_SIZE_Y)):
            __main(
                browser_path=browser_path,
                user_agent=user_agent,
                headless=False,
                target_url=target_url,
                attempts=attempts,
                test_connection=test_connection,
                app_key=app_key,
            )
    except test.Results as e:
        e.handle_result()

if __name__ == "__main__":
    main()
