FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PIP_NO_CACHE_DIR=1

WORKDIR /app

RUN groupadd \
        --gid 10001 \
        django \
    && useradd \
        --uid 10001 \
        --gid django \
        --create-home \
        django

COPY requirements.txt /app/requirements.txt

RUN python -m pip install \
    --upgrade pip \
    && python -m pip install \
    --requirement /app/requirements.txt

COPY --chown=django:django . /app

RUN mkdir -p \
        /app/data \
        /app/staticfiles \
    && chown -R django:django \
        /app/data \
        /app/staticfiles \
    && chmod +x \
        /app/docker-entrypoint.sh

USER django

EXPOSE 8000

HEALTHCHECK \
    --interval=10s \
    --timeout=5s \
    --start-period=60s \
    --retries=12 \
    CMD python -c \
    "import urllib.request; urllib.request.urlopen('http://localhost:8000/', timeout=4)"

ENTRYPOINT ["/app/docker-entrypoint.sh"]

CMD ["gunicorn", "config.wsgi:application", "--bind", "0.0.0.0:8000", "--workers", "3", "--threads", "2", "--timeout", "120", "--access-logfile", "-", "--error-logfile", "-"]