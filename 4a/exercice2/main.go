package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func effectuerTache(id int, wg *sync.WaitGroup) {
	// Garantit que wg.Done() sera appelé quand la goroutine se termine
	defer wg.Done()

	fmt.Printf("Goroutine %d: Début de la tâche...\n", id)

	duree := time.Duration(50+rand.Intn(451)) * time.Millisecond
	time.Sleep(duree)

	fmt.Printf("Goroutine %d: Tâche terminée.\n", id)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1) // Indique qu'une goroutine supplémentaire est attendue
		go effectuerTache(i, &wg)
	}

	fmt.Println("Toutes les goroutines lancées.")

	// Bloque main jusqu'à ce que toutes les goroutines appellent wg.Done()
	wg.Wait()

	fmt.Println("Toutes les goroutines ont terminé leur exécution.")
}
