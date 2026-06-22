package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// travailleur lit des IDs de tâche depuis le canal taches, les traite,
// et envoie les résultats sur le canal resultats.
// La boucle s'arrête automatiquement quand taches est fermé (for range).
func travailleur(id int, taches <-chan int, resultats chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for idTache := range taches {
		fmt.Printf("Travailleur %d: traitement de la tâche %d...\n", id, idTache)

		duree := time.Duration(50+rand.Intn(451)) * time.Millisecond
		time.Sleep(duree)

		resultat := fmt.Sprintf("Travailleur %d: tâche %d terminée.", id, idTache)
		resultats <- resultat
	}

	fmt.Printf("Travailleur %d: plus de tâches, je m'arrête.\n", id)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	const nbTravailleurs = 3
	const nbTaches = 10

	taches := make(chan int, nbTaches)
	resultats := make(chan string, nbTaches)

	var wg sync.WaitGroup

	// Lance les travailleurs
	for i := 1; i <= nbTravailleurs; i++ {
		wg.Add(1)
		go travailleur(i, taches, resultats, &wg)
	}

	// Envoie toutes les tâches dans le canal
	for i := 1; i <= nbTaches; i++ {
		taches <- i
	}
	// Ferme taches : signale aux travailleurs qu'il n'y a plus de travail
	close(taches)

	// Attend la fin de tous les travailleurs dans une goroutine séparée
	// pour pouvoir fermer resultats après
	go func() {
		wg.Wait()
		close(resultats)
	}()

	// Collecte et affiche tous les résultats
	for msg := range resultats {
		fmt.Println(msg)
	}

	fmt.Println("Toutes les tâches ont été traitées.")
}
