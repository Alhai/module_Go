# JOURNAL_IA.md — Usage de l'IA

## Contexte

J'ai utilisé ChatGPT (GPT-4o) ponctuellement pendant ce TP, principalement pour débloquer des questions de syntaxe Go que je ne maîtrisais pas encore, et pour avoir un deuxième avis sur certaines décisions d'architecture. Je suis resté l'architecte du projet — l'IA n'a jamais écrit de bloc de code que j'ai collé sans le lire.

---

## Partie 0 — Setup

J'ai initialisé le module et l'arborescence moi-même, je connais `go mod init` depuis les TPs précédents. J'ai demandé à l'IA un `.gitignore` adapté à Go parce que je ne me souvenais plus des fichiers à exclure (binaires, `vendor/`, etc.). Elle m'a sorti une liste, j'en ai retiré quelques entrées inutiles pour ce projet (`vendor/`, `*.test`) et gardé l'essentiel.

---

## Partie 1 — Domaine

J'ai rédigé les structs `CheckResult` et `Batch` moi-même en suivant le contrat JSON du sujet. Ce qui m'a posé problème : les struct tags. J'avais mis `json:"status_code"` sans `omitempty` au départ, et la réponse JSON incluait `"status_code": 0` pour les URLs en erreur, ce qui ne correspondait pas à l'exemple du sujet. J'ai demandé à l'IA comment omettre un champ entier quand il vaut zéro — elle m'a confirmé `omitempty`, ce que j'aurais pu trouver dans la doc mais ça m'a fait gagner 5 minutes.

Pour `NewSummary`, j'ai voulu partir sur une `map[bool]int` pour compter up/down, l'IA m'a dit que c'était valide mais que c'était moins lisible qu'une simple boucle avec deux compteurs. Elle avait raison, j'ai gardé la version simple.

Les interfaces `Checker` et `Store` viennent directement du sujet, je les ai recopiées.

Pour `ErrBatchNotFound`, j'avais d'abord fait `var ErrBatchNotFound = fmt.Errorf("batch not found")` puis l'IA a signalé que `errors.New` est plus idiomatique pour les sentinelles (pas de formatting). J'ai changé.

---

## Partie 2 — Store

Le store en mémoire, je l'ai écrit seul. J'avais oublié d'utiliser un `RWMutex` au départ (juste un `Mutex`), l'IA m'a expliqué la différence : `RLock` pour les lectures concurrentes, `Lock` uniquement pour les écritures. J'ai refactorisé `Get` pour utiliser `RLock`. Ça a du sens vu qu'on fait beaucoup plus de lectures que d'écritures dans ce service.

---

## Partie 3 — Checker

Le `HTTPChecker` : j'ai d'abord utilisé `http.Get(url)` sans context. L'IA m'a rappelé qu'il fallait `http.NewRequestWithContext` pour que l'annulation via `context` fonctionne. J'ai corrigé.

Le `MockChecker` : idée personnelle de le mettre dans un fichier non-`_test.go` pour pouvoir l'importer depuis les tests d'autres packages. L'IA a confirmé que c'est la bonne pratique en Go quand le mock doit être partagé.

---

## Partie 4 — Pool concurrent

C'est la partie où j'ai le plus consulté l'IA, parce que les goroutines et les channels c'est ce qui me pose le plus problème.

**Premier jet (rejeté)** : j'avais lancé une goroutine par URL sans borne — `for _, url := range urls { go ... }`. L'IA m'a dit que c'était un anti-pattern explicitement interdit dans le sujet. J'ai refait avec un pool.

**Deuxième jet** : j'avais un channel de résultats non bufferisé. Le test bloquait. L'IA m'a expliqué pourquoi : si tous les workers essaient d'envoyer un résultat en même temps et que le collecteur n'est pas encore prêt, deadlock. J'ai mis le channel en bufferisé de taille `len(urls)`.

**Fermeture du channel** : j'avais mis `close(results)` après la boucle `for range work` dans chaque worker — ce qui faisait que plusieurs workers essayaient de fermer le même channel. Panic. L'IA m'a montré le pattern `go func() { wg.Wait(); close(results) }()`. C'est élégant et ça règle le problème.

**Annulation du context** : le `select { case <-ctx.Done(): ... default: }` en tête de boucle, j'ai eu du mal à comprendre pourquoi on ne faisait pas juste laisser `Check` échouer via le context. L'IA a expliqué : si le context est déjà annulé avant même d'entrer dans `Check`, on évite de créer inutilement un `context.WithTimeout` enfant. C'est une micro-optimisation mais c'est plus propre.

---

## Partie 5 — API REST

**Routing** : je ne savais pas que Go 1.22 permettait `mux.HandleFunc("GET /v1/checks/{id}", ...)` directement dans la stdlib. J'avais commencé à installer `gorilla/mux` mais l'IA m'a dit que ce n'était plus nécessaire depuis 1.22. J'ai viré la dépendance.

**Validation** : j'ai écrit `validateRequest` moi-même. J'avais oublié de vérifier que les URLs commençaient bien par `http://` ou `https://`, l'IA me l'a signalé en relecture.

**generateID** : j'avais utilisé `math/rand` avec `rand.Seed(time.Now().UnixNano())`. L'IA m'a conseillé `crypto/rand` pour éviter les collisions dans un contexte serveur. Le format `b_%x` sur 3 bytes donne exactement 6 chars hex, ce qui correspond à l'exemple du sujet (`b_4f3c1a`).

**Middleware de logging** : j'ai voulu loguer le `batch_id` dans la réponse. L'IA a proposé de passer par un context value pour le transmettre du handler au middleware — mais après réflexion, j'ai trouvé ça trop complexe pour le gain. J'ai choisi de loguer le `batch_id` directement dans le handler avec `h.logger.Info(...)`. Plus simple, plus lisible.

**RecoveryMiddleware** : première version écrasait les headers si `WriteHeader` avait déjà été appelé. L'IA m'a expliqué que `http.ResponseWriter` ne permet pas de réécrire le statut une fois envoyé. J'ai réorganisé le `defer` pour que le recover intervienne avant toute écriture.

---

## Partie 6 — Tests

J'ai écrit les tests table-driven pour `NewSummary` moi-même, c'est un pattern qu'on a vu en cours. Pour les tests de handlers avec `httptest`, j'ai demandé à l'IA un exemple minimal — elle m'a montré `httptest.NewRequest` + `httptest.NewRecorder`. J'ai adapté.

Un bug que j'ai trouvé seul : dans `TestGetBatchNotFound`, j'appelais le handler directement sans passer par le mux, donc `r.PathValue("id")` retournait une chaîne vide. J'ai dû ajouter `req.SetPathValue("id", "b_nope")` explicitement. Ça m'a pris 20 minutes à comprendre.

Le `go test -race ./...` est passé propre dès le deuxième essai (après avoir corrigé la fermeture du channel de résultats).

---

## Bilan

L'IA m'a surtout servi à :
- Débloquer rapidement des questions de syntaxe ou d'API stdlib que j'aurais trouvées dans la doc en cherchant plus longtemps
- Valider des choix d'architecture (channel bufferisé, RWMutex, crypto/rand)
- Relire du code et signaler des oublis (omitempty, context sur les requêtes HTTP)

Ce que j'ai rejeté ou corrigé :
- L'idée d'une `map[bool]int` pour le résumé (gardé la boucle simple)
- L'utilisation de `math/rand` pour les IDs
- Le passage du `batch_id` via context value dans le middleware (trop complexe)
- Un premier design de pool sans channel bufferisé
