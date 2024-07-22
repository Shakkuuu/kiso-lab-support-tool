FROM golang:1.22.5

RUN apt-get update && \
    apt-get install -y python3 python3-pip python3-venv

RUN python3 -m venv /opt/venv

RUN /opt/venv/bin/pip install PyMuPDF Pillow

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

ENV VIRTUAL_ENV=/opt/venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

CMD ["/bin/sh", "-c", "./main -user $USER_ENV -password $PASSWORD_ENV -port $PORT_ENV"]
