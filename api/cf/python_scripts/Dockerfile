FROM python:3.12-slim-bullseye

COPY . /app
WORKDIR /app

ENV KJHJASON_CF_SANDBOX=1

# Install some other dependencies
RUN apt-get update && \
    apt-get install -y xvfb gnupg2 curl wget unzip ca-certificates

# Install Google Chrome browser
RUN curl -sS -o - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo "deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list && \
    apt-get -y update && \
    apt-get -y install google-chrome-stable && \
    rm -rf /var/lib/apt/lists/*

# Install python dependencies
RUN pip install --no-cache-dir fastapi -r requirements.txt 

EXPOSE 8080

CMD ["python", "api_wrapper.py"]
