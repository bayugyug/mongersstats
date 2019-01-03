## mongersstats

* [x] Simple golang utility library on safe queueing of stats


### Install

```sh

go get -u -v github.com/bayugyug/mongersstats

```

### Mini-How-To


```go

package main

import (
        "fmt"
        "time"

        stats "github.com/bayugyug/mongersstats"
       )


func main() {

        q, _ := stats.NewQ()
        i := 0
        for {
                time.Sleep(1 * time.Millisecond)

                //increment by 1
                q.Incr("STATS")

                //increment by 1.0
                q.FloatIncr("DECIMAL::STATS")

                //reload & print
                i++
                if i == 50 {
                        fmt.Println(q.JSONify())
                        q.Reload()
                }

                //done
                if i >= 100 {
                        break
                }
        }
        //show again :-)
        fmt.Println(q.Stringify())


}
```





### Reference




### License

[MIT](https://bayugyug.mit-license.org/)

