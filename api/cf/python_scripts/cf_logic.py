# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass Cloudflare protection using DrissionPage."""

import time
import logging
import functools
from DrissionPage import (
    ChromiumPage,
)
from DrissionPage.common import (
    Actions,
)

CF_WRAPPER_XPATH = ".cf-turnstile-wrapper"
LOGGER_NAME = "cf_bypass"

def configure_logger(log_path: str) -> None:
    logger = logging.getLogger(LOGGER_NAME)
    logger.setLevel(logging.INFO)

    file_handler = logging.FileHandler(log_path, encoding="utf-8")
    file_handler.setLevel(logging.INFO)

    formatter = logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")
    file_handler.setFormatter(formatter)

    logger.addHandler(file_handler)

def get_logger() -> logging.Logger:
    return logging.getLogger(LOGGER_NAME)

@functools.lru_cache
def get_base_url(url: str) -> str:
    try:
        url = url.split("/", maxsplit=3)
        url = "/".join(url[:3])
    except IndexError:
        pass
    return url

def __is_bypassed(page: ChromiumPage, logger: logging.Logger) -> None:
    logger.info("Checking if bypassed...")
    if not page.wait.doc_loaded(timeout=30):
        logger.error("Page failed to load, retrying...")
        return False

    for packet in page.listen.steps(count=1, timeout=4.5):
        if packet.response.status != 403:
            return True
        return False
    # The only edge case is where the 
    # target url redirects to another url
    # which shouldn't happen since it's kinda weird.
    logger.warning("No packets received with listener...")

    # # old code for fanbox.cc
    # if target_url == "https://www.fanbox.cc":
    #     return page.wait.ele_displayed(r"xpath://a[@href='/']",timeout=2.5)

    # Note: doesn't work for custom pages
    html_lang = page.ele("tag:html", timeout=1.5).attr("lang")
    if html_lang == "en" or html_lang == "en-US":
        title = page.title.lower()
        return "just a moment" not in title

    # in the event the user's system is not set to en-US
    return not page.wait.ele_displayed(CF_WRAPPER_XPATH, timeout=2.5)

def __bypass_logic(page: ChromiumPage, logger: logging.Logger) -> None:
    if not page.wait.ele_displayed(CF_WRAPPER_XPATH, timeout=1.5):
        logger.error(f"{CF_WRAPPER_XPATH} Element not found at {page.url} retrying...")
        logger.info(f"HTML Content for reference:\n{page.html}\n")
        return

    time.sleep(1.5)
    actions = Actions(page)

    # Move mouse to the CF Wrapper
    actions.move_to(CF_WRAPPER_XPATH, duration=0.75)

    # Tries to move 120px to the left
    # from current position in a human-like manner
    actions.left(130).wait(0.15, 0.45).right(10)

    # left click and hold for 
    # 0.01~0.15 seconds (randomised) before releasing
    actions.hold().wait(0.01, 0.15).release()

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
    #     cf_wrapper: ShadowRoot | None = page.ele(CF_WRAPPER_XPATH).shadow_root
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

def __bypass(page: ChromiumPage, attempts: int, logger: logging.Logger) -> bool:
    if __is_bypassed(page, logger):
        logger.info("Bypassed!")
        return True

    logger.info("Challenge page detected...")
    time.sleep(4)
    logger.info("trying to bypass...")
    __bypass_logic(page, logger)
    time.sleep(4)

    logger.error(f"Failed to bypass after {attempts} attempts")
    return False

def bypass_cf(page: ChromiumPage, attempts: int, logger: logging.Logger = get_logger()) -> bool:
    time.sleep(4)
    if attempts > 0:
        for _ in range(attempts):
            if __bypass(page, attempts, logger):
                return True
        return False

    while True:
        if __bypass(page, attempts, logger):
            return True
