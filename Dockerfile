## Using multi stage build feature.
FROM golang:1.16 as builder

COPY . /src/

WORKDIR /src/

RUN CGO_ENABLED=0 go build -buildmode=exe -o cloner  /src/cmd/cloner

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /src/cloner .

USER nonroot:nonroot

ENTRYPOINT ["/cloner"]
