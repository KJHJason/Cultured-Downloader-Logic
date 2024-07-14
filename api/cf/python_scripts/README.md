# CF Solver by KJHJason

[![License: GPLv3](https://img.shields.io/badge/license-GPLv3-blue)](https://opensource.org/license/gpl-3-0)

Repository: [KJHJason/Cultured-Downloader-Logic](https://github.com/KJHJason/Cultured-Downloader-Logic/tree/main/api/cf/python_scripts)

Simple script to bypass CF protection using [DrissionPage](https://github.com/g1879/DrissionPage).

Notes:

- Logic based on one of the repositories by [sarperavci](https://github.com/sarperavci).
- This script will NOT work if your IP address already has a bad reputation!

## Virtual Displays

Although the script does work with the `--headless=new` option, it is recommended to run without headless mode to reduce the risk of getting detected.

Hence, this script uses [pyvirtualdisplay](https://github.com/ponty/pyvirtualdisplay) which uses `xvfb` under the hood to create a virtual display and run Chrome (non-headless mode) on it without showing the Chrome GUI on your main display.

Caveats of using virtual displays:

- For Windows, the script will ignore the `--virtual-display` flag as it requires `xvfb` to create a virtual display.
  - Note: `xvfb` is not available for Windows.
  - However, you can run the script using the provided Docker image instead for virtual display using `xvfb` .
- For Linux, you will need to install `xvfb`.
  - `sudo apt-get install xvfb`
  - Run `xvfb-run` to check if `xvfb` is installed and working.
- For MacOS, you will need to install [XQuartz](https://www.xquartz.org/) for X11.
  - `brew install XQuartz`
  - You might need to add `/usr/X11/bin` to your PATH.
  - Run `which xvfb` to check if `xvfb` is installed.
  - If you're facing issues, look up for a fix on StackOverflow as I don't have a MacOS device to test this.
  - Alternatively, you can run the script using the provided Docker image.
