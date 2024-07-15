"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import os
import platform

__version__ = "0.1.0"
PLATFORM_NAME = platform.system()
IS_LINUX = PLATFORM_NAME == "Linux"
IS_MACOS = PLATFORM_NAME == "Darwin"
IS_UNIX = IS_LINUX or IS_MACOS
IS_DOCKER = os.getenv("KJHJASON_CF_DOCKER") == "1"

WINDOW_SIZE_X = 1920
WINDOW_SIZE_Y = 1080
USING_PY_AUTO_GUI_KEY = "USING_PY_AUTO_GUI"

ARGS_BOOLEAN_CHOICE = ("true", "True", "1", "false", "False", "0",)
OS_CHOICES = ("linux", "darwin", "windows",) # from Golang's runtime.GOOS

DEFAULT_TARGET_URL = "https://nopecha.com/demo/cloudflare"
CF_WRAPPER_XPATH = ".cf-turnstile-wrapper"
LOGGER_NAME = "cf_bypass"

EXTENSIONS_DIR = "./extensions"
NAVIGATOR_EXT_DIR = EXTENSIONS_DIR + "/Navigator"
NAVIGATOR_JS_PATH = NAVIGATOR_EXT_DIR + "/navigator.js"
