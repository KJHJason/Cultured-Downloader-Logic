FROM python:3.12-slim-bullseye

COPY . /app
WORKDIR /app

# Install some other dependencies
RUN apt update && \
    apt install -y wget curl procps vim xvfb unzip

# Install Chrome browser binary
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
    dpkg -i google-chrome-stable_current_amd64.deb || apt --fix-broken install -y && \
    rm -rf google-chrome-stable_current_amd64.deb

# Install python dependencies
RUN pip install --no-cache-dir fastapi -r requirements.txt 

EXPOSE 8080

CMD ["python", "api_wrapper.py"]
