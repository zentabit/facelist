FROM ubuntu:20.04
LABEL NAME=aboutme-dl

RUN apt-get update && apt-get install -y python3 python3-pip && apt-get clean

WORKDIR /app

COPY * /app/
RUN pip3 install -r requirements.txt

CMD ["python3","convert.py","/data"]
