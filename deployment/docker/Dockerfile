FROM  alpine/git:2.43.0 as gitclone-build

WORKDIR /clone
RUN git clone https://github.com/esatus/didcomm-rust

FROM rust:1.78 as rust-build

COPY --from=gitclone-build /clone/didcomm-rust /src
WORKDIR /src

#RUN build the rust library
RUN cd uniffi && cargo build --release

FROM golang:1.21 as go-build

# Copies your source code into the go directory
COPY . /src

# set work directory
WORKDIR /src

# copy lib from rust build
RUN mkdir -p didcomm/lib
COPY --from=rust-build /src/uniffi/target/release/libdidcomm_uniffi.* didcomm/lib/
COPY ./cmd/api/database /src/database/
#RUN go build to build the application with the name app
RUN GOOS=linux GOARCH=amd64 go build -o /src/app ./cmd/api/

EXPOSE 9090

ENV LD_LIBRARY_PATH=/src/didcomm/lib

CMD ["/src/app"]

# FROM scratch
# # copy lib and dependencies (finde dependencies with: ldd <executable>)
# COPY --from=go-build /src/didcomm/lib /bin/lib/
# COPY --from=go-build /lib/x86_64-linux-gnu/libgcc_s.so.1 /lib/x86_64-linux-gnu/libgcc_s.so.1
# COPY --from=go-build /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
# COPY --from=go-build /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
# COPY --from=go-build /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
# COPY --from=go-build /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
# COPY --from=go-build /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2

# # add curl for healthcheck
# COPY --from=ghcr.io/tarampampam/curl:8.6.0 /bin/curl /bin/curl

# # copy config file
# COPY --from=go-build /src/cmd/api/config.yaml /config.yaml
# # copy binary
# COPY --from=go-build /src/app /bin/app
# # set enviroment variable for lib
# ENV LD_LIBRARY_PATH=/bin/lib

# EXPOSE 9090:9090

# HEALTHCHECK --interval=5s --timeout=2s --retries=2 --start-period=2s CMD ["curl", "--fail", "http://127.0.0.1:9090/health"]

# CMD ["/bin/app"]
