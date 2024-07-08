# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass Cloudflare protection using DrissionPage.

Note: Logic based on https://github.com/sarperavci/CloudflareBypassForScraping
"""

import time
import logging
from DrissionPage import (
    ChromiumPage,
    errors,
)

FANBOX_URL = "https://www.fanbox.cc"

# used to work but cf decided to add #shadow-root (closed) to the iframe.
# though not sure about custom pages which fanbox uses.
IFRAME_XPATH = "xpath://div/iframe"
IFRAME_DIV_WRAPPER_XPATH = "#turnstile-wrapper"

def __is_bypassed(driver: ChromiumPage, target_url: str) -> None:
    logging.info("Checking if bypassed...")
    if target_url == FANBOX_URL:
        return driver.wait.ele_displayed(r"xpath://a[@href='/']",timeout=1.5)

    # Note: doesn't work for custom pages
    html_lang = driver.ele("xpath://html", timeout=1.5).attr("lang")
    if html_lang == "en-US":
        title = driver.title.lower()
        return "just a moment" not in title

    # in the event the user's system is not set to en-US
    return not driver.wait.ele_displayed(IFRAME_DIV_WRAPPER_XPATH, timeout=2.5)

def __bypass_logic(driver: ChromiumPage, target_url: str) -> None:
    selector = ""
    if target_url == FANBOX_URL and driver.wait.ele_displayed(IFRAME_XPATH, timeout=1.5):
        selector = IFRAME_XPATH

    if selector == "":
        if not driver.wait.ele_displayed(IFRAME_DIV_WRAPPER_XPATH, timeout=1.5):
            logging.error(f"{IFRAME_DIV_WRAPPER_XPATH} Element not found, retrying...")
            return
        selector = IFRAME_DIV_WRAPPER_XPATH

    time.sleep(1.5)
    if selector == IFRAME_DIV_WRAPPER_XPATH:
        # Since the IFRAME_DIV_WRAPPER_XPATH is guaranteed to exist due to the check above,
        # we can safely use the ele() without wrapping it in a try-except block.
        driver.ele(selector, timeout=2.5).click()
        return

    try:
        iframe = driver(selector)
        checkbox = iframe.ele("xpath://input[@type='checkbox']", timeout=2.5)
        checkbox.click()
    except errors.ElementLostError:
        logging.error("Checkbox element not found, retrying...")
        return

def bypass_cf(driver: ChromiumPage, target_url: str) -> None:
    while not __is_bypassed(driver, target_url):
        logging.info("Challenge page detected, trying to bypass...")
        time.sleep(5)
        __bypass_logic(driver, target_url)
        time.sleep(3)
