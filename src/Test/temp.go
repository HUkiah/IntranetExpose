package main

import (
	"fmt"
	"time"
)

func main() {

	fmt.Println("Main...")

	Limiter := make(chan time.Time, 0)

	request := make(chan int, 0)
	response := make(chan int, 0)

	//1号
	go func() {
		fmt.Println("1号...")

		defer func() {
			fmt.Println("1号 exit...")
		}()

		for t := range time.Tick(time.Second * 2) {
			Limiter <- t
		}

	}()

	//2号
	go func() {

		fmt.Println("2号...")

		defer func() {
			fmt.Println("2号 exit..")
		}()

		var n = 1
		for {

			<-Limiter
			if n < 5 {
				request <- n
				n++
			} else {
				response <- 6
				break
			}

		}

	}()

	//3号
	go func() {

		fmt.Println("3号...")

		defer func() {
			fmt.Println("3号 exit...")
		}()

		for {

			fmt.Println("select for...")
			select {

			case <-request:
				fmt.Println("case request..")
			case <-response:
				fmt.Println("case response..")
				goto Top
			}

		}
	Top:
	}()

	defer func() {
		fmt.Println("Main exit")
	}()

	time.Sleep(30 * time.Second)

}
