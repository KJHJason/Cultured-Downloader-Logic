"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import shutil
import typing
import random
import functools
import subprocess

import _logger
import constants
from .errors import (
    handle_err,
)

import pyautogui

@functools.lru_cache(maxsize=1)
def get_base_url(url: str) -> str:
    try:
        url = url.split("/", maxsplit=3)
        url = "/".join(url[:3])
    except IndexError:
        pass
    return url

def check_for_xvfb() -> bool:
    if shutil.which("xvfb-run") is not None:
        return True

    try:
        subprocess.run(["Xvfb", "-help"], check=True)
    except subprocess.CalledProcessError:
        _logger.get_logger().warning("xvfb-run not found, ignoring --virtual-display flag...")
        return False
    return True

def check_container(app_key: str) -> None | typing.NoReturn:
    # Mainly just to make it harder to run the script in a container.
    if constants.IS_DOCKER and app_key != "fzN9Hvkb9s+mwPGCDd5YFnLiqKx8WhZfWoZE5nZC":
        handle_err("Failed to connect to browser...")

def reset_mouse_position() -> None:
    offset = 100
    if constants.IS_DOCKER:
        screen_h = random.randint(offset, constants.WINDOW_SIZE_Y - offset)
        screen_w = random.randint(offset, constants.WINDOW_SIZE_X - offset)
        pyautogui.moveTo(screen_w, screen_h)
    else:
        screen_size = pyautogui.size()
        screen_h = random.randint(offset, screen_size.height - offset)
        screen_w = random.randint(offset, screen_size.width - offset)
        pyautogui.moveTo(screen_w, screen_h)
