ver=0.0.3
for p in linux_amd64 linux_arm64 darwin_amd64 darwin_arm64; do
  nix store prefetch-file \
    --hash-type sha256 \
    "https://github.com/wneessen/localweather/releases/download/v${ver}/localweather_${ver}_${p}.tar.gz"
done
