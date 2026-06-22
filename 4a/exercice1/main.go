package main

import (
	"fmt"
	"math/rand"
	"time"
)

func effectuerTache(id int) {
	fmt.Printf("Goroutine %d: Début de la tâche...\n", id)

	// Simule un travail entre 50 et 500ms
	duree := time.Duration(50+rand.Intn(451)) * time.Millisecond
	time.Sleep(duree)

	fmt.Printf("Goroutine %d: Tâche terminée.\n", id)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	for i := 1; i <= 5; i++ {
		go effectuerTache(i)
	}

	fmt.Println("Toutes les goroutines lancées.")

	// Sans synchronisation, main se termine ici immédiatement.
	// Les goroutines n'ont pas le temps de finir leur travail.
}
