"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import time

import _logger
import constants

from DrissionPage import (
    ChromiumPage,
)

def __is_bypassed(page: ChromiumPage, target_url: str) -> None:
    logger = _logger.get_logger()
    logger.info("Checking if bypassed...")
    if not page.wait.doc_loaded(timeout=30):
        logger.error("Page failed to load, retrying...")
        return False

    logger.info("Page loaded successfully, listening to packets...")
    for packet in page.listen.steps(count=1, timeout=4.5):
        if packet.response.status != 403:
            return True
        return False

    # This should only happen if the bypass wasn't successful.
    # Otherwise, one edge case is where the 
    # target url redirects to another url
    # which shouldn't happen since it's kinda weird.
    logger.warning("No packets received with listener...")

    if target_url == "https://www.fanbox.cc":
        return page.wait.ele_displayed(r"xpath://a[@href='/']",timeout=3.5)

    # Note: doesn't work for custom pages
    html_lang = page.ele("tag:html", timeout=1.5).attr("lang")
    if html_lang == "en" or html_lang == "en-US":
        title = page.title.lower()
        return "just a moment" not in title

    # in the event the user's system is not set to en-US
    return not page.wait.ele_displayed(constants.CF_WRAPPER_XPATH, timeout=3.5)

def __bypass_logic(page: ChromiumPage) -> None:
    logger = _logger.get_logger()
    if not page.wait.ele_displayed(constants.CF_WRAPPER_XPATH, timeout=2.5):
        logger.error(f"{constants.CF_WRAPPER_XPATH} Element not found at {page.url} retrying...")
        logger.info(f"HTML Content for reference:\n{page.html}\n")
        return

    # sleep to wait for the checkbox to appear
    time.sleep(3)

    actions = page.actions

    # Move mouse to the CF Wrapper
    actions.move_to(constants.CF_WRAPPER_XPATH, duration=0.75)

    # Tries to move 120px to the left
    # from current position in a human-like manner
    actions.left(130).wait(0.15, 0.45).right(10)

    # left click and hold for 
    # 0.01~0.15 seconds (randomised) before releasing
    actions.hold().wait(0.01, 0.15).release()

    # sleep for the cf to verify the click
    time.sleep(4.5)

    # # Old code below that uses .click() on the checkbox element
    # from DrissionPage import (
    #     ChromiumPage,
    #     errors as drission_errors,
    # )
    # from DrissionPage._elements.chromium_element import (
    #     ChromiumElement,
    #     ShadowRoot,
    # )
    # try:
    #     cf_wrapper: ShadowRoot | None = page.ele(constants.CF_WRAPPER_XPATH).shadow_root
    #     if cf_wrapper is None:
    #         logger.error("cf wrapper ShadowRoot not found, retrying...")
    #         return
    # except drission_errors.ElementNotFoundError:
    #     logger.error("cf wrapper element not found, retrying...")
    #     return

    # try:
    #     iframe: ChromiumElement = cf_wrapper.ele("tag:iframe", timeout=2.5)
    # except drission_errors.ElementNotFoundError:
    #     logger.error("iframe element not found, retrying...")
    #     return

    # try:
    #     iframe.ele("tag:input", timeout=2.5).click()
    # except drission_errors.ElementNotFoundError:
    #     logger.error("checkbox element not found, retrying...")
    #     return

def __bypass(page: ChromiumPage, target_url: str) -> bool:
    logger = _logger.get_logger()
    if __is_bypassed(page, target_url):
        logger.info("Bypassed!")
        return True

    logger.info("Challenge page detected...")
    logger.info("trying to bypass...")
    __bypass_logic(page)
    return False

def bypass_cf(page: ChromiumPage, attempts: int, target_url: str) -> bool:
    logger = _logger.get_logger()
    if attempts > 0:
        for _ in range(attempts):
            if __bypass(page, target_url):
                return True
        logger.error(f"Failed to bypass after {attempts} attempts")
        return False

    while True:
        if __bypass(page, target_url):
            return True
