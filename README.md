## mongersstats

* [x] Simple golang utility library on safe queueing of stats in a concurrent kind of application (concurrency/multi-threaded)


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


var q *stats.Stats

func main() {

        q, _ = stats.NewQ()

        isReady := make(chan bool, 1)
        go multiRun(isReady)

        isMore := make(chan bool, 1)
        go multiMore(isMore)

        <-isReady
        <-isMore

        //Note: 
        //      try to wait a bit the channel
        //      but in real app, this shouldnt be needed
        //
        time.Sleep(1 * time.Millisecond) 
        fmt.Println(q.Stringify())
}

func multiRun(ready chan bool) {

        i := 0
        for {
                i++

                //increment by 1
                q.Incr("STATS")

                //increment by 1.0
                q.FloatIncr("DECIMAL::STATS")

                //done
                if i >= 100 {
                        break
                }
        }
        ready <- true
}

func multiMore(ready chan bool) {

        i := 0
        for {
                i++

                //increment by 1
                q.Incr("MORE::STATS")

                //increment by 1.0
                q.FloatIncr("MORE::DECIMAL::STATS")

                //done
                if i >= 500 {
                        break
                }
        }
        ready <- true
}

```





### Reference




### License

[MIT](https://bayugyug.mit-license.org/)

