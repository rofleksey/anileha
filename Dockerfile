FROM golang:1.20-alpine AS apiBuilder
WORKDIR /opt
COPY . /opt/
RUN apk add --no-cache alpine-sdk
RUN go mod download
RUN go build -o ./anileha

FROM node:18 AS frontBuilder
WORKDIR /opt
COPY ./frontend/ /opt/
RUN npm i && npm run build

FROM alpine
WORKDIR /opt
RUN apk update && apk add --no-cache curl ca-certificates ffmpeg
COPY --from=apiBuilder /opt/anileha /opt/anileha
COPY --from=frontBuilder /opt/dist /opt/frontend/dist
EXPOSE 8080
CMD [ "./anileha" ]
