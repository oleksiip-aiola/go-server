package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type engine interface {
	milesLeft() uint8
}

func canMakeIt(e engine, miles uint8) {
	if miles <= e.milesLeft() {
		fmt.Println("We can make it!")
	} else {
		fmt.Println("We can't make it!")
	}
}

func square(itsquared *[5]float64) *[5]float64 {
	fmt.Printf("memory loc %p\n", itsquared)

	for i := 0; i < len(itsquared); i++ {
		itsquared[i] = itsquared[i] * itsquared[i]
	}

	return itsquared
}

// func dbCall(i int) {
// 	// Simulate DB call delay
// 	var delay float32 = 2000
// 	time.Sleep(time.Duration(delay) * time.Millisecond)
// 	save(dbData[i])
// 	log()
// 	wg.Done()
// }

// func save(res string) {
// 	rwm.Lock()
// 	defer rwm.Unlock()
// 	result = append(result, res)
// }

// func log() {
// 	rwm.RLock()
// 	defer rwm.RUnlock()
// }

var rwm = sync.RWMutex{}
var m = sync.Mutex{}
var wg = sync.WaitGroup{}
var dbData = []string{"id1", "id2", "id3", "id4", "id5"}
var result = []string{}

// func main() {

// 	// var p *int32 = new(int32)
// 	// var i int32 = 5
// 	// fmt.Println(*p, p, &i)

// 	// *p = 10
// 	// fmt.Println(*p, p)

// 	// p = &i
// 	// fmt.Println(*p, p, i, &i)

// 	// *p = 15

// 	// fmt.Println(*p, p, i, &i)

// 	// var pointerToElectricEngineStruct *electricEngine = &electricEngine{}
// 	// fmt.Println(pointerToElectricEngineStruct, &electricEngine{}, *pointerToElectricEngineStruct, electricEngine{})

// 	// var tosquare [5]float64 = [5]float64{1, 2, 3, 4, 6}
// 	// fmt.Printf("memory loc %p\n", &tosquare)
// 	// var squared = *square(&tosquare)
// 	// fmt.Println(squared)
// 	// squared[1] = 100

// 	// fmt.Println(squared)
// 	// fmt.Println(tosquare)

// 	// t0 := time.Now()
// 	// for i := 0; i < 10; i++ {
// 	// 	wg.Add(1)
// 	// 	go dbCall(2)
// 	// }

// 	// wg.Wait()

// 	// fmt.Printf("\nTotal execution time: %v\n", time.Since(t0))
// }

var MAX_CHICKEN_PRICE float32 = 5

func checkChickenPrice(source string, chickenChannel chan string) {
chickenLoop:
	for {
		time.Sleep(time.Second * 1)
		var price = rand.Float32() * 20
		if price <= MAX_CHICKEN_PRICE {
			chickenChannel <- source
			break chickenLoop
		}
	}
}

func sendMessage(channel chan string) {
	fmt.Printf("\nFound source with lowest price :%s", "yes")
}

func sumAnything[T int | float64 | string](slice []T) T {
	var sum T

	for _, v := range slice {
		sum += v
	}

	return sum
}

type gasEngine struct {
	mpg     uint8
	gallons uint8
}

type electricEngine struct {
	mpkwh uint8
	kwh   uint8
}

type car[T gasEngine | electricEngine] struct {
	model        string
	yearProduced string
	engine       T
	secondEngine T
}

func main() {
	// intarr := []int{1, 2, 3, 4, 5}
	// floatarr := []float64{1, 23, 4, 5, 6.5}
	// stringarr := []string{"asd", "dga", "ababbb"}

	// fmt.Println(sumAnything(intarr), sumAnything(floatarr), sumAnything(stringarr))

	// electricCar := car[gasEngine]{
	// 	model:        "Tesla",
	// 	yearProduced: "1995",
	// 	engine:       gasEngine{mpg: 1, gallons: 2},
	// 	secondEngine: gasEngine{mpg: 2, gallons: 3},
	// }

	// y := getValue()

	// fmt.Println(*y / 2)

	// strslice := []int{1, 2, 3, 4, 5, 6}

	// strsliced := strslice[1:3]
	// fmt.Println(strsliced)

	fmt.Println(fibo(12))
	fmt.Println(fibAppend(12))
	fmt.Println(recursiveFibo(12, nil))

}

func getValue() *int {
	x := 2

	return &x

}

func fibo(val int) []int {
	slice := make([]int, val)
	for i := range slice {
		if i < 2 {
			slice[i] = i
			continue
		}
		slice[i] = slice[i-1] + slice[i-2]
	}

	return slice
}

func fibAppend(val int) []int {
	slice := []int{0, 1}

	for len(slice) < val {
		slice = append(slice, slice[len(slice)-1]+slice[len(slice)-2])
	}

	return slice
}

func recursiveFibo[T []int](val int, sum T) []int {
	slice := sum

	if sum == nil {
		sum = []int{0, 1}
	}

	if len(sum) == val {
		return sum
	}
	slice = append(slice, slice[len(slice)-1]+slice[len(slice)-2])
	return recursiveFibo(val, slice)

}
