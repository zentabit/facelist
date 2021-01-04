FROM ubuntu:20.04
LABEL NAME=aboutme-dl

RUN apt update && apt install python3 python3-pip && apt clean

WORKDIR /app

COPY * /app/
RUN pip install -r requirements.txt

CMD ["python3","convert.py","/data"]
