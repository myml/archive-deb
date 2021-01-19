# 这是什么

这是一个简单的 go 模块，提供类似 archive/tar 的 API，用于读写 Debian deb

# 例子

```go
func main() {
	debFile := "./test.deb"
	f, _ := os.Open(debFile)
	defer f.Close()
	r := deb.NewReader(f)
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if strings.HasPrefix(header.Name, "DEBIAN/control") {
			data, _ := ioutil.ReadAll(r)
			log.Println("control file", string(data))
		}
		if strings.HasPrefix(header.Name, "data")  && !header.FileInfo().IsDir() {
			log.Println("data file", header.Name, header.Size)
		}
	}
}
```
