# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass Cloudflare protection using DrissionPage.

Note: Logic based on https://github.com/sarperavci/CloudflareBypassForScraping
"""

import time
import logging
from DrissionPage import (
    ChromiumPage,
)
from DrissionPage.common import (
    Actions,
)

CF_WRAPPER_XPATH = ".cf-turnstile-wrapper"

def __is_bypassed(driver: ChromiumPage, target_url: str) -> None:
    logging.info("Checking if bypassed...")
    if target_url == "https://www.fanbox.cc":
        return driver.wait.ele_displayed(r"xpath://a[@href='/']",timeout=2.5)

    # Note: doesn't work for custom pages
    html_lang = driver.ele("tag:html", timeout=1.5).attr("lang")
    if html_lang == "en" or html_lang == "en-US":
        title = driver.title.lower()
        return "just a moment" not in title

    # in the event the user's system is not set to en-US
    return not driver.wait.ele_displayed(CF_WRAPPER_XPATH, timeout=2.5)

def __bypass_logic(driver: ChromiumPage) -> None:
    if not driver.wait.ele_displayed(CF_WRAPPER_XPATH, timeout=1.5):
        logging.error(f"{CF_WRAPPER_XPATH} Element not found at {driver.url} retrying...")
        logging.info(f"HTML Content for reference:\n{driver.html}\n")
        return

    time.sleep(1.5)
    actions = Actions(driver)

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
    #     cf_wrapper: ShadowRoot | None = driver.ele(CF_WRAPPER_XPATH).shadow_root
    #     if cf_wrapper is None:
    #         logging.error("cf wrapper ShadowRoot not found, retrying...")
    #         return
    # except drission_errors.ElementNotFoundError:
    #     logging.error("cf wrapper element not found, retrying...")
    #     return

    # try:
    #     iframe: ChromiumElement = cf_wrapper.ele("tag:iframe", timeout=2.5)
    # except drission_errors.ElementNotFoundError:
    #     logging.error("iframe element not found, retrying...")
    #     return

    # try:
    #     iframe.ele("tag:input", timeout=2.5).click()
    # except drission_errors.ElementNotFoundError:
    #     logging.error("checkbox element not found, retrying...")
    #     return

def bypass_cf(driver: ChromiumPage, target_url: str) -> None:
    while not __is_bypassed(driver, target_url):
        logging.info("Challenge page detected...")
        time.sleep(4)
        logging.info("trying to bypass...")
        __bypass_logic(driver)
        time.sleep(4)
