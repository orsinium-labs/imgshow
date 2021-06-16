# imgshow

Dead simple Go library to open a window with an image.

Only Linux (X server) is supported.

## Installation

```bash
go get github.com/orsinium-labs/imgshow
```

## Usage

```go
stream, _ := os.Open("image.png")
img, _ := png.Decode(stream)
imgshow.Show(img)
```

See [_examples](./examples/) directory for more examples.

## Standing on the shoulders of giants

The code is based on [goiv](https://github.com/gen2brain/goiv) library which is an image viewer written on Go. Unfortunately, the project doesn't provide an API, so `imgshow` was born out of it.
