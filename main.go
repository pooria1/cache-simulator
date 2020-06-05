package main

import (
	"cache-simulator/core"
	"fmt"
	"strconv"
)

func main() {

	var blockSize int
	var cacheType int
	var assoc int
	var wh string
	var wm string
	var size int

	fmt.Scanf("%d%s", &blockSize)
	fmt.Scanf("%d%s", &cacheType)
	fmt.Scanf("%d%s", &assoc)
	fmt.Scanf("%s%s", &wh)
	fmt.Scanf("%s", &wm)
	fmt.Scanf("%d", &size)

	// Options
	options := &core.Options{
		Size:       uint64(size),
		Assoc:      assoc,
		BlockSize:  uint64(blockSize),
		MissPolicy: wm,
		HitPolicy:  wh,
		Debug:      false, // set true for debugging

	}
	// Start Cache
	cache := &core.Cache{Options: options}
	cache.Init()

	for {
		var operation string
		var address string
		fmt.Scanf("%s", &operation)
		if operation == "" {
			break
		}
		fmt.Scanf("%s%s", &address)
		fmt.Println("here is it :", operation, address)
		op, _ := strconv.Atoi(operation)
		err := cache.Execute(op, address)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("***CACHE SETTINGS***")
	if cacheType == 1 {
		fmt.Println("Split I- D-cache")
	} else {
		fmt.Println("Unified I- D-cache")
	}
	fmt.Println("Size:", cache.Options.Size)
	fmt.Println("Associativity:", cache.Options.Assoc)
	fmt.Println("Block Size:", cache.Options.BlockSize)
	fmt.Println("Write Policy:", cache.Options.HitPolicy)
	fmt.Println("Allocation Policy:", cache.Options.MissPolicy)

	fmt.Println("\n***CACHE STATISTICS***")
	fmt.Println("INSTRUCTIONS")
	fmt.Println("accesses: ????????")
	fmt.Println("misses: ????????")
	fmt.Print("miss rate: ????????")
	fmt.Println(" (hit rate: ????????)")
	fmt.Println("replaces: ????????")

	fmt.Println("DATA")
	fmt.Println("INSTRUCTIONS")
	fmt.Println("accesses: ????????")
	fmt.Println("misses:", cache.Stats.Misses)
	fmt.Print("miss rate:", cache.Stats.CalculateMissRate())
	fmt.Printf(" (hit rate:%f)\n", 1-cache.Stats.CalculateMissRate())
	fmt.Print("replaces: ????????")

}
