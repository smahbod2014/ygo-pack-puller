package main

import (
	"log"
	"math/rand"
	"smahbod2014/ygo-pack-puller/api"
	"time"
)

func main() {
	seed := time.Now().Unix()
	rand.Seed(seed)
	log.Println("Initializing with seed", seed)

	api.Start()
}
