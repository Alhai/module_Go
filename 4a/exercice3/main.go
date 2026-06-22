package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func effectuerTache(id int, wg *sync.WaitGroup, resultChan chan string) {
	defer wg.Done()

	fmt.Printf("Goroutine %d: Début de la tâche...\n", id)

	duree := time.Duration(50+rand.Intn(451)) * time.Millisecond
	time.Sleep(duree)

	fmt.Printf("Goroutine %d: Tâche terminée.\n", id)

	// Envoie le résultat sur le canal avant que wg.Done() ne soit appelé
	resultChan <- fmt.Sprintf("Goroutine %d a terminé avec succès.", id)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	// Canal bufferisé : évite le blocage si main n'a pas encore commencé à lire
	resultChan := make(chan string, 5)

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go effectuerTache(i, &wg, resultChan)
	}

	fmt.Println("Toutes les goroutines lancées.")

	// Attend la fin de toutes les goroutines dans une goroutine séparée
	// pour pouvoir fermer le canal ensuite
	go func() {
		wg.Wait()
		// Ferme le canal pour signaler qu'aucune autre donnée ne sera envoyée
		close(resultChan)
	}()

	// Lit tous les résultats jusqu'à la fermeture du canal
	for msg := range resultChan {
		fmt.Println("Résultat reçu:", msg)
	}

	fmt.Println("Tous les résultats ont été traités.")
}
