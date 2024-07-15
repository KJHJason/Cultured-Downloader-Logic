# Author: KJHJason <contact@kjhjason.com>.
# License: GNU GPL v3.

"""Simple script to bypass CF protection using DrissionPage.
"""

import sys
import typing

import _logger

class Results(Exception):
    def __init__(self, success: bool) -> None:
        self.success = success

    def __str__(self) -> str:
        return f"Test {'succeeded' if self.success else 'failed'}"

    def handle_result(self) -> typing.NoReturn:
        logger = _logger.get_logger()
        if self.success:
            logger.info("Test succeeded")
            print("Test succeeded")
            sys.exit(0)
        else:
            logger.error("Test failed")
            print("Test failed")
            sys.exit(1)
