FROM python:latest

COPY . /app
WORKDIR /app

# Install chrome and other dependencies like xvfb
RUN apt-get update && \
    apt-get install -y wget xvfb \
        libgl1-mesa-dri chromium chromium-common chromium-driver && \
    wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
    apt-get install -y ./google-chrome-stable_current_amd64.deb && \
    rm ./google-chrome-stable_current_amd64.deb && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install depends
RUN pip install --no-cache-dir -r requirements.txt && \
    pip install fastapi pydantic uvicorn

EXPOSE 8080

CMD ["python", "api_wrapper.py"]
