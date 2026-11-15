# Local Home Assistant add-on of the GMT fork.
# The image is the upstream beta image (launcher, FFmpeg, runtime) with the
# bridge binary replaced by the fork's build. The binary is cross-compiled on
# the PC (scripts/build-gmt.sh), so nothing is compiled on the Pi.
FROM ghcr.io/afrouper/steinel-cam-bridge:1.5.2-beta.11
COPY --chmod=0755 steinel-bridge /app/steinel-bridge
