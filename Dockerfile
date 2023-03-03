#build stage
FROM golang:1.16.2-alpine as builder 
WORKDIR /gophie
COPY go.mod ./
COPY go.sum ./
RUN go mod download 
COPY . .
RUN go build -o /gophie
EXPOSE 3000

# deploy stage
FROM alpine:latest
WORKDIR /gophie 
COPY --from=builder /gophie /gophie 
CMD [ "./gophie","api","-p","3000" ]
