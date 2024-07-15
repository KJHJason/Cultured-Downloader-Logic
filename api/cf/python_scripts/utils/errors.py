"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import sys
import typing

import _logger

def handle_err(msg: str) -> typing.NoReturn:
    print(msg)
    _logger.get_logger().error(msg)
    sys.exit(1)
