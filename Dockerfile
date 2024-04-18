FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-tailscale"]
COPY baton-tailscale /