FROM ubuntu:latest as base

RUN apt-get update
RUN apt-get -y install curl wget git
RUN apt-get -y install gcc make
RUN apt-get -y install golang-1.23
RUN apt-get -y install ca-certificates

ENV PATH /usr/lib/go-1.23/bin:$PATH

FROM  alpine/git:2.43.0 as gitclone-build

WORKDIR /clone
RUN git clone https://github.com/esatus/didcomm-rust

FROM rust:1.73 as rust-build

COPY --from=gitclone-build /clone/didcomm-rust /src
WORKDIR /src

#RUN build the rust library
RUN cd uniffi && cargo build --release

FROM base

# Copies your source code into the go directory
COPY . /src

# set work directory
WORKDIR /src

# copy lib from rust build
RUN mkdir -p didcomm/lib
COPY --from=rust-build /src/uniffi/target/release/libdidcomm_uniffi.* didcomm/lib/
RUN echo ls
COPY ./cmd/api/database /src/database/
#RUN go build to build the application with the name app
#RUN GOOS=linux GOARCH=amd64 go build -o /src/app ./cmd/api/
#
#EXPOSE 9090
#
#ENV LD_LIBRARY_PATH=/src/didcomm/lib

#CMD ["/src/app"]