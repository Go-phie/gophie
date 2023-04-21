#build stage
FROM golang:1.20-alpine as builder 
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download 
COPY . .
RUN go build -o /gophie

# deploy stage
FROM alpine:latest
COPY --from=builder /gophie /gophie 
CMD [ "./gophie","api","-p","3000" ]
EXPOSE 3000
