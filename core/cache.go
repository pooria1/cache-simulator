package core

import (
	"fmt"
	"math"
	"strconv"
)

// Operations enums
const (
	DataStore = 0
	DataLoad  = 1
)

// Tag is an alias for the tab (variable size byte array)
type Tag uint64

// Cache represents a literal cache
type Cache struct {
	// Cache Info
	Options *Options
	Stats   *Statistics

	// Cache Info
	NumberOfSets uint64 // Size / (Granularity * Associativity)
	SetNumber    uint64 // Set Number: log2 (Number of Sets)
	BlockOffset  uint64 // Block Offset: log2 (Block Size)
	TagBits      uint64 // Number of remaining bits
	TotalBits    uint64 // Total number of index bites

	// Cache Components
	TagArray   [][]Tag // Array of tags for lookup purposes
	DirtyArray [][]bool
}

// Options represents the init options for a cache
type Options struct {
	Size       uint64
	Assoc      int
	BlockSize  uint64
	HitPolicy  string
	MissPolicy string
	Debug      bool
}

// Init initializes the cache based on the embedded cache options
func (c *Cache) Init() error {
	// Check
	if c.Options == nil {
		return fmt.Errorf("unable to instantiate cache: options not configured")
	}
	if c.Options.HitPolicy == "wt" {
		c.Options.HitPolicy = "WRITE BACK"
	} else {
		c.Options.HitPolicy = "WRITE THROUGH"
	}
	if c.Options.MissPolicy == "wa" {
		c.Options.MissPolicy = "WRITE ALLOCATE"
	} else {
		c.Options.MissPolicy = "NO WRITE ALLOCATE"
	}
	o := c.Options

	// Compute cache info
	c.TotalBits = uint64(math.Log2(float64(o.Size)))
	c.NumberOfSets = o.Size / (o.BlockSize * uint64(o.Assoc))
	c.SetNumber = uint64(math.Log2(float64(c.NumberOfSets)))
	c.BlockOffset = uint64(math.Log2(float64(o.BlockSize)))
	c.TagBits = c.TotalBits - (c.SetNumber + c.BlockOffset)

	// Build Cache Components
	// Note that tag array is a 2d array of bytes
	// structured as follows:
	c.TagArray = make([][]Tag, c.NumberOfSets)
	for i := range c.TagArray {
		c.TagArray[i] = make([]Tag, 0, o.Assoc)
	}

	if o.HitPolicy == "WRITE BACK" {
		c.DirtyArray = make([][]bool, c.NumberOfSets)
		for i := range c.DirtyArray {
			c.DirtyArray[i] = make([]bool, 0, o.Assoc)
		}
	}

	c.Stats = new(Statistics)
	return nil
}

// Execute executes a cache operation based on the cache options
func (c *Cache) Execute(op int, address string) error {
	// Parse Address: string --> []byte
	// Ignore block offset
	tag, set, _, err := c.Parse(address)
	if err != nil {
		return err
	}

	// Handle Hit/Miss
	hit, index := c.Lookup(set, tag)

	if op == DataStore {
		// Write
		err := c.Write(set, index, tag, hit)
		return err
	} else if op == DataLoad {
		err := c.Read(set, index, tag, hit)
		return err
	} else {
		// BAD OP
		return fmt.Errorf("bad cache operation code: %v", op)
	}
}

// Parse converts an address into the tag, set number, and block offset
func (c *Cache) Parse(address string) (Tag, uint64, uint64, error) {

	intAddr, err := strconv.ParseUint(address, 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	addr := intAddr

	offset := addr
	set := addr >> (c.BlockOffset)
	tag := set >> (c.SetNumber)

	// Clear the tag bits
	for i := uint64(0); i <= 64-(c.SetNumber); i++ {
		set = selectiveClear(set, 64-i)
	}

	// Clear the tag and cache index bits
	for i := uint64(0); i <= 64-(c.BlockOffset); i++ {
		offset = selectiveClear(offset, 64-i)
	}
	return Tag(tag), set, offset, nil
}

func selectiveClear(n uint64, pos uint64) uint64 {
	if n&(1<<pos) > 0 {
		n = n &^ (1 << pos)
	}
	return n
}

// Lookup checks the tag array to determines if a block is in the cache
// and modifies stats accordingly
func (c *Cache) Lookup(set uint64, tag Tag) (bool, uint64) {

	for index := range c.TagArray[set] {
		if c.TagArray[set][index] == tag {
			c.Stats.Hits++
			return true, uint64(index)
		}
	}

	c.Stats.Misses++
	return false, 0
}

// Write performs a cache write
func (c *Cache) Write(set uint64, index uint64, tag Tag, hit bool) error {
	// Writes always modify the cache block
	// If WriteBack is on, the modify flag sets the dirty array
	// If WriteTrough is on, the modify flag does nothing
	return c.Replace(set, index, tag, hit, true)
}

// Read performs a cache read
func (c *Cache) Read(set uint64, index uint64, tag Tag, hit bool) error {
	// Reads never modify a block
	if hit == false && len(c.TagArray[set]) == c.Options.Assoc {
		c.Stats.Replaces++
	}
	err := c.Replace(set, index, tag, hit, false)
	return err
}

// Replace executes a replacement using the proper replacement policy
func (c *Cache) Replace(set uint64, index uint64, tag Tag, hit bool, modify bool) error {
	var eviction bool

	// Retrieve from memory
	if !hit {
		c.Stats.Reads++
	}

	eviction = c.LRU(set, index, tag, hit, modify)

	if c.Options.HitPolicy == "WRITE BACK" {
		if eviction {
			c.Stats.Writes++
		}
	} else if c.Options.HitPolicy == "WRITE THROUGH" && modify {
		c.Stats.Writes++
	}

	return nil
}

// LRU simulates LRU replacement
func (c *Cache) LRU(set uint64, index uint64, tag Tag, hit bool, modify bool) bool {
	var eviction bool

	// If the tag is currently in the cache
	// Make sure to perform an LRU replacement
	if hit {

		// Check LRU position
		if index == 0 {
			c.TagArray[set] = c.TagArray[set][1:]
		} else if int(index) == len(c.TagArray[set])-1 {
			c.TagArray[set] = c.TagArray[set][:index]
		} else {
			// cut center
			c.TagArray[set] = append(c.TagArray[set][:index], c.TagArray[set][index+1:]...)
		}

		// Append
		c.TagArray[set] = append(c.TagArray[set], tag)

		if c.Options.HitPolicy == "WRITE BACK" {
			var b = c.DirtyArray[set][index]
			if index == 0 {
				// cut head
				c.DirtyArray[set] = c.DirtyArray[set][1:]
			} else if int(index) == len(c.DirtyArray[set])-1 {
				c.DirtyArray[set] = c.DirtyArray[set][:index]
			} else {
				// cut center
				c.DirtyArray[set] = append(c.DirtyArray[set][:index], c.DirtyArray[set][index+1:]...)
			}

			if modify {
				c.DirtyArray[set] = append(c.DirtyArray[set], modify)
			} else {
				c.DirtyArray[set] = append(c.DirtyArray[set], b)
			}
		}
		return false
	}

	// Insert: Check if cache is full
	// Insert into MRU position
	if len(c.TagArray[set]) == c.Options.Assoc {
		// Evict the LRU position
		c.TagArray[set][0] = 0
		c.TagArray[set] = c.TagArray[set][1:]

		if c.Options.HitPolicy == "WRITE BACK" {
			eviction = c.DirtyArray[set][0]
			c.DirtyArray[set] = c.DirtyArray[set][1:]
		}
	}

	// Insert the new tag
	c.TagArray[set] = append(c.TagArray[set], tag)

	// Set dirty bit
	if c.Options.HitPolicy == "WRITE BACK" {
		c.DirtyArray[set] = append(c.DirtyArray[set], modify)
	}
	return eviction
}
