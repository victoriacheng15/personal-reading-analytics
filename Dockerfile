FROM python:3.10-slim-bullseye

WORKDIR /usr/app

COPY requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt

COPY src ./src
COPY .env ./
COPY credentials.json ./

WORKDIR /usr/app/src

CMD ["python", "main.py"]
