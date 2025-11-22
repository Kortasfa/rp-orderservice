FROM gcr.io/distroless/static-debian12
ADD bin/order /app/order
ENTRYPOINT ["/app/order"]
