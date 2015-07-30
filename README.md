# AdBlock

`adblock` is a package for working with Adblock Plus filter rules. It can parse Adblock Plus filters and match URLs against them.

# Install

```
$ go get -u github.com/bluele/adblock
```

# Examples

```go
package main

import (
  "fmt"
  "github.com/bluele/adblock"
)

func main() {
  ab, err := adblock.NewRules([]string{
    "||google.com/ads/$domain=black.example.com",
    "||example.com/",
    "@@||example.com/white/",
  }, nil)
  if err != nil {
    panic(err)
  }

  fmt.Println("http://google.com/ads/app.js", ab.ShouldBlock("http://google.com/ads/app.js", map[string]interface{}{
    "domain": "white.example.com",
  }))
  fmt.Println("http://google.com/ads/app.js", ab.ShouldBlock("http://google.com/ads/app.js", map[string]interface{}{
    "domain": "black.example.com",
  }))
  fmt.Println("http://example.com/", ab.ShouldBlock("http://example.com/", nil))
  fmt.Println("http://example.com/white/", ab.ShouldBlock("http://example.com/white/", nil))
}
```

Result:

```
http://google.com/ads/app.js false
http://google.com/ads/app.js true
http://example.com/ true
http://example.com/white/ true
```

# Dependency

* [golang-pkg-pcre](https://github.com/glenn-brown/golang-pkg-pcre) location: https://github.com/bluele/adblock/tree/master/regexp/pcre
* libpcre++

# Author

**Jun Kimura**

* <http://github.com/bluele>
* <junkxdev@gmail.com>