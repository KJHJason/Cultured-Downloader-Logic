"""
@Author   : KJHJason
@Contact  : contact@kjhjason.com
@Copyright: (c) 2024 by KJHJason. All Rights Reserved.
@License  : GNU GPL v3
"""

import logging
import functools

import constants

def configure_logger(log_path: str) -> None:
    logger = logging.getLogger(constants.LOGGER_NAME)
    logger.setLevel(logging.INFO)

    file_handler = logging.FileHandler(log_path, encoding="utf-8")
    file_handler.setLevel(logging.INFO)
    if file_handler.stream.tell() > 0:
        # add a newline for better readability
        file_handler.stream.write("\n")

    formatter = logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")
    file_handler.setFormatter(formatter)

    logger.addHandler(file_handler)

@functools.cache
def get_logger() -> logging.Logger:
    return logging.getLogger(constants.LOGGER_NAME)
