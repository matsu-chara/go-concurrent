package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func lexical() {
	chanOwner := func() <-chan int {
		results := make(chan int, 5)
		go func() {
			defer close(results)
			for i := 0; i <= 5; i++ {
				results <- i
			}
		}()
		return results
	}

	consumer := func(results <-chan int) {
		for result := range results {
			fmt.Printf("Received: %d\n", result)
		}
		fmt.Println("Done receiving!")
	}

	results := chanOwner()
	consumer(results)
}

func forSelect() {
	doWork := func(
		done <-chan interface{},
		strings <-chan string,
	) <-chan interface{} {
		terminated := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(terminated)
			for {
				select {
				case s := <-strings:
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()
		return terminated
	}

	done := make(chan interface{})
	terminated := doWork(done, nil)

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("canceling doWork goroutine...")
		close(done)
	}()

	<-terminated
	fmt.Println(done)
}

func randomStream() {
	newRandStream := func(done <-chan interface{}) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStraeam closure exited.")
			defer close(randStream)
			for {
				select {
				case randStream <- rand.Int():
				case <-done:
					return
				}
			}
		}()
		return randStream
	}

	done := make(chan interface{})
	randomStream := newRandStream(done)
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randomStream)
	}
	close(done)
}

func orPattern() {
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}
		orDone := make(chan interface{})
		go func() {
			defer close(orDone)
			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}
}

func erroHandlingBad() {
	checkStatus := func(
		done <-chan interface{},
		urls ...string,
	) <-chan *http.Response {
		responses := make(chan *http.Response)
		go func() {
			defer close(responses)
			for _, url := range urls {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					continue
				}
				select {
				case <-done:
					return
				case responses <- resp:
				}
			}
		}()
		return responses
	}

	done := make(chan interface{})
	defer close(done)

	urls := []string{"https://www.google.com", "https://badhost"}
	for response := range checkStatus(done, urls...) {
		fmt.Println("Response: %v\n", response.Status)
	}
}

func erroHandlingGood() {
	type Result struct {
		Error    error
		Response *http.Response
	}
	checkStatus := func(
		done <-chan interface{},
		urls ...string,
	) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)
			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)

	urls := []string{"https://www.google.com", "https://badhost"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v", result.Error)
			continue
		}
		fmt.Println("Response: %v\n", result.Response.Status)
	}
}

func pipelines() {
	generator := func(done <-chan interface{}, integers ...int) <-chan int {
		intCh := make(chan int, len(integers))
		go func() {
			defer close(intCh)
			for _,i := range integers {
				select {
				case <- done :
					return
				case intCh <- i:
				}
			}
		}()
		return intCh
	}

	multiply := func(
		done <-chan interface{},
		intCh <-chan int,
		multiplier int,
	) <-chan int {
		multipliedCh := make(chan int)
		go func() {
			defer close(multipliedCh)
			for i:= range intCh {
				select {
				case <- done :
					return
				case multipliedCh <- i * multiplier:
				}
			}
		}()
		return multipliedCh
	}

	add := func(
		done <-chan interface{},
		intCh <-chan int,
		additive int,
	) <-chan int {
		addedCh := make(chan int)
		go func() {
			defer close(addedCh)
			for i:= range intCh {
				select {
				case <- done :
					return
				case addedCh <- i + additive:
				}
			}
		}()
		return addedCh
	}

	done := make(chan interface{})
	defer close(done)

	intCh := generator(done, 1,2,3,4)
	pipeline := multiply(done, add(done, multiply(done, intCh, 2), 1), 2)

	for v:= range pipeline {
		fmt.Println(v)
	}
}

func main() {
	// lexical()
	// forSelect()
}
