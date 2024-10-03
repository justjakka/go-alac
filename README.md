# go-alac

[libalac](https://github.com/mikebrady/alac) CGO bindings for encoding/decoding.

## Original repo
https://git.gammaspectra.live/S.O.N.G/go-alac

## Requirements
### [libalac](https://github.com/mikebrady/alac)
```shell
git clone --depth 1 https://git.gammaspectra.live/S.O.N.G/alac.git
cd alac
autoreconf -fi
./configure --prefix /usr
make -j$(nproc)
sudo make install
```
