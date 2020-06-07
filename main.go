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
	var dataCacheSize int
	var instructionCacheSize int

	fmt.Scanf("%d%s", &blockSize)
	fmt.Scanf("%d%s", &cacheType)
	fmt.Scanf("%d%s", &assoc)
	fmt.Scanf("%s%s", &wh)
	fmt.Scanf("%s", &wm)
	fmt.Scanf("%d", &dataCacheSize)
	instructionCacheSize = dataCacheSize
	if cacheType == 1 {
		fmt.Scanf("%s")
		fmt.Scanf("%d", &instructionCacheSize)
	}

	// Options
	dataCacheOptions := &core.Options{
		Size:       uint64(dataCacheSize),
		Assoc:      assoc,
		BlockSize:  uint64(blockSize),
		HitPolicy:  wh,
		MissPolicy: wm,
	}
	instructionCacheOptions := &core.Options{
		Size:       uint64(instructionCacheSize),
		Assoc:      assoc,
		BlockSize:  uint64(blockSize),
		HitPolicy:  wh,
		MissPolicy: wm,
	}
	// Start Cache
	dataCache := &core.Cache{Options: dataCacheOptions}
	instructionCache := &core.Cache{Options: instructionCacheOptions}
	dataCache.Init()
	instructionCache.Init()

	for {
		var operation string
		var address string
		fmt.Scanf("%s", &operation)
		if operation == "" {
			break
		}
		fmt.Scanf("%s%s", &address)
		op, _ := strconv.Atoi(operation)
		var err error
		if op == 2 {
			if cacheType == 1 {
				err = instructionCache.Execute(core.DataLoad, address)
				instructionCache.Stats.Accesses++
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			err = instructionCache.Execute(core.DataLoad, address)
			instructionCache.Stats.Accesses++
			if err != nil {
				fmt.Println(err)
			}
			for k, _ := range instructionCache.TagArray {
				dataCache.DirtyArray[k] = instructionCache.DirtyArray[k]
				dataCache.TagArray[k] = instructionCache.TagArray[k]
			}
			continue
		}
		err = dataCache.Execute(op, address)
		dataCache.Stats.Accesses++
		if err != nil {
			fmt.Println(err)
		}
		for k, _ := range instructionCache.TagArray {
			instructionCache.DirtyArray[k] = dataCache.DirtyArray[k]
			instructionCache.TagArray[k] = dataCache.TagArray[k]
		}
	}

	// Output
	fmt.Println("***CACHE SETTINGS***")
	if cacheType == 1 {
		fmt.Println("Split I- D-dataCache")
		fmt.Println("I-cache size:", instructionCache.Options.Size)
		fmt.Println("D-cache size:", dataCache.Options.Size)
	} else {
		fmt.Println("Unified I- D-dataCache")
		fmt.Println("Size:", dataCache.Options.Size)
	}
	fmt.Println("Associativity:", dataCache.Options.Assoc)
	fmt.Println("Block Size:", dataCache.Options.BlockSize)
	fmt.Println("Write Policy:", dataCache.Options.HitPolicy)
	fmt.Println("Allocation Policy:", dataCache.Options.MissPolicy)

	fmt.Println("\n***CACHE STATISTICS***")
	fmt.Println("INSTRUCTIONS")
	fmt.Println("accesses:", instructionCache.Stats.Accesses)
	fmt.Println("misses:", instructionCache.Stats.Misses)
	fmt.Printf("miss rate: %.4f", instructionCache.Stats.CalculateMissRate())
	if instructionCache.Stats.Accesses == 0 {
		fmt.Println(" (hit rate: 0.0000)")
	} else {
		fmt.Printf(" (hit rate: %.4f)\n", 1-instructionCache.Stats.CalculateMissRate())
	}
	fmt.Println("replaces:", instructionCache.Stats.Replaces)

	fmt.Println("DATA")
	fmt.Println("INSTRUCTIONS")
	fmt.Println("accesses:", dataCache.Stats.Accesses)
	fmt.Println("misses:", dataCache.Stats.Misses)
	fmt.Print("miss rate: ", dataCache.Stats.CalculateMissRate())
	fmt.Printf(" (hit rate: %.4f)\n", 1-dataCache.Stats.CalculateMissRate())
	fmt.Print("replaces: ", dataCache.Stats.Replaces)

}
