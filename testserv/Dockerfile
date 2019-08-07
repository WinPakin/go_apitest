FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY . .
# COPY ../ackpb/ack.pb.go ./apitest/ackpb


# Get Dependencies
RUN go get github.com/WinPakin/ackpb
RUN go get google.golang.org/grpc
RUN go get github.com/gonum/stat
RUN go get github.com/gorilla/mux

# Build the Go app
RUN go build -o main .

# Expose port 5002 to the outside world
EXPOSE 5002

# Command to run the executable
CMD ["./main"]