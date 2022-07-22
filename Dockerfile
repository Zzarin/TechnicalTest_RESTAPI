FROM golang:latest
#Therefore, instead of creating our own base image, we’ll use the official Go image that already has all the tools and packages to compile and run a Go application.
#When we have used that FROM command, we told Docker to include in our image all the functionality from the @golang:latest@ image

#create a directory inside the image that we are building. This also instructs Docker to use this directory as the default destination for all subsequent commands.
#This way we do not have to type out full file paths but can use relative paths based on this directory.
WORKDIR /app

#COPY go.mod ./
#COPY go.sum ./

#COPY command takes two parameters. The first parameter tells Docker what files you want to copy into the image. The last parameter tells Docker where you want that file to be copied to.
# We’ll copy the go.mod and go.sum file into our project directory /app which, owing to our use of WORKDIR, is the current directory (.) inside the image.
COPY ./ ./

#Now that we have the module files inside the Docker image that we are building, we can use the RUN command to execute the command go mod download
#Go modules will be installed into a directory inside the image.
RUN go mod download

#This COPY command uses a wildcard to copy all files with .go extension located in the current directory on the host
#(the directory where the Dockerfile is located) into the current directory inside the image.
RUN go build -o /api

EXPOSE 4057

CMD ["/api"] 