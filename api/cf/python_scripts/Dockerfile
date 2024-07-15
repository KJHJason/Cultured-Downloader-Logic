ARG PYTHON_VERSION=3.12
FROM python:${PYTHON_VERSION}-slim-bullseye

COPY . /app
WORKDIR /app

ENV KJHJASON_CF_DOCKER=1
ENV KJHJASON_CF_SANDBOX=1
ENV KJHJASON_CF_ADD_NAV_EXT=1
ENV KJHJASON_CF_SET_MAX_WINDOW=1

# Install python dependencies
RUN python -m pip install --upgrade pip
RUN pip install --no-cache-dir -r requirements.txt 

# Install dependencies for xvfb and for installing Google Chrome from source
RUN apt-get update && \
    apt-get install -y xvfb gnupg2 curl wget unzip ca-certificates

# Install Google Chrome browser
RUN curl -sS -o - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo "deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list && \
    apt-get -y update && \
    apt-get -y install google-chrome-stable && \
    rm -rf /var/lib/apt/lists/*

# Test the script to see if it runs and also to generate .pyc files
RUN python main.py --version

ENTRYPOINT [ "python", "main.py" ]
