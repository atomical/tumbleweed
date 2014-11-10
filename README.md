### Tumbleweed

Backup images from a tumblr account.

### Usage

```
tumbleweed arizonanature
```


### Cross compiling tumbleweed for distribution

```
# windows
cd /usr/local/go/src; GOOS=windows GOARCH=386 CGO_ENABLED=0 ./make.bash --no-clean
GOOS=windows GOARCH=386 go build -o tumbleweed.exe tumbleweed.go
```