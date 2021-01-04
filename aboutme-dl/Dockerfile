FROM ubuntu:20.04

WORKDIR /app

COPY * /app/
RUN pip install -r requirements.txt

CMD ["python3","convert.py","/data"]
