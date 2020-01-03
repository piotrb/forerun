FROM ubuntu

RUN apt-get update && apt-get install -yy python

ARG TAG

COPY forerun /usr/bin/forerun
COPY sleeper.sh .
COPY inner.py .
COPY Procfile .
