"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import logging

class CfError(Exception):
    def __init__(self, msg: str) -> None:
        self.msg = msg

    def __str__(self) -> str:
        return self.msg

def handle_err(msg: str, logger: logging.Logger) -> None:
    print(msg)
    logger.error(msg)
    raise CfError(msg)
