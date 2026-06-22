# DESIGN.md — URLWatch

## Découpage en packages

Le projet suit une architecture hexagonale légère :

- **`domain`** : types métier (`CheckResult`, `Batch`, `Summary`) et interfaces (`Checker`, `Store`). Aucune dépendance vers les autres packages internes. C'est la seule couche que `pool`, `store`, `checker` et `api` partagent.
- **`checker`** : implémentation HTTP du `Checker`. Le `MockChecker` vit ici (fichier non `_test.go`) pour être importable par les tests des packages `pool` et `api`, sans créer de dépendances circulaires.
- **`pool`** : cœur concurrent — connaît `domain.Checker` et `domain.CheckResult`, rien d'autre.
- **`store`** : implémentation mémoire du `Store`, indépendante de `api` et de `pool`.
- **`api`** : connait `domain`, `pool` et `store` (via les interfaces). Responsable du décodage JSON, de la validation, du routing et du logging.
- **`cmd/urlwatch`** : assemblage uniquement — crée les instances concrètes et les injecte.

Les frontières d'interface se situent entre `api` ↔ `checker` et `api` ↔ `store`, ce qui permet de tester les handlers sans réseau ni persistance réelle.

## Modèle de concurrence

**Worker pool borné** : toutes les URLs sont pré-chargées dans un channel bufferisé (taille = `len(urls)`) avant le démarrage des workers. Ce choix évite de bloquer l'émetteur et garantit que la fermeture du channel est faite une seule fois, avant le lancement des goroutines. Le channel est directement itéré via `range work` dans chaque worker.

Le channel de résultats est aussi bufferisé (taille = `len(urls)`) pour que les workers ne bloquent jamais sur l'émission d'un résultat, même si le collecteur principal est lent.

**WaitGroup** : chaque worker appelle `wg.Done()` en `defer`. Une goroutine dédiée attend `wg.Wait()` puis ferme le channel de résultats — ce qui débloque le `range results` du collecteur. Cette séquence garantit l'absence de deadlock et de fuite de goroutine.

**Échecs partiels** : chaque URL produit exactement un `CheckResult`, qu'elle soit en succès ou en erreur. Pas d'abandon global sur première erreur ; le résumé reflète le bilan réel.

**Context** : deux niveaux de timeout — le context du batch (timeout HTTP par URL via `context.WithTimeout(ctx, timeoutMs)`) et le context de la requête HTTP serveur (annulation si le client déconnecte). À l'annulation, le `select { case <-ctx.Done(): ... }` en tête de boucle des workers retourne un résultat d'erreur immédiatement pour les URLs non encore traitées.

## Fuites de goroutines

Risques identifiés et mesures :

| Risque | Mesure |
|---|---|
| Worker bloqué sur `results <-` si le collecteur n'est pas là | Channel bufferisé (taille `len(urls)`) |
| `wg.Wait()` goroutine qui ne se termine pas | Impossible : workers se terminent quand `work` est vide (channel fermé avant démarrage) |
| HTTP roundtrip qui bloque indéfiniment | `context.WithTimeout` par URL + annulation propagée depuis le context parent |

La suite de tests passe sous `go test -race` sans erreur.

## Stratégie d'erreurs

- **Sentinelle** `domain.ErrBatchNotFound` : retournée par `Store.Get`, wrappée via `fmt.Errorf("...: %w", ...)`. La couche API détecte avec `errors.Is` et renvoie `404`.
- **Type d'erreur** `domain.ValidationError` : porte le nom du champ fautif, détectée avec `errors.As` dans le handler. Renvoie `400 invalid_request`.
- **Wrapping systématique** : les erreurs traversent les couches enveloppées, préservant la chaîne pour le débogage sans exposer les détails internes au client JSON.

## Pourquoi Go ici plutôt que Java, Python ou Rust

1. **Goroutines et channels natifs** : le modèle CSP de Go rend le worker pool lisible en ~30 lignes sans librairie externe. En Java, il faudrait `ExecutorService`, `CompletableFuture` et une gestion explicite des threadpools. En Python, l'asyncio ou le multiprocessing ajoutent de la complexité et le GIL reste une contrainte.

2. **Binaire statique et démarrage instantané** : le service compile en un seul binaire sans dépendances runtime, ce qui simplifie le déploiement (Docker scratch, Lambda). Rust ferait aussi bien sur ce point, mais le cycle `borrow checker` aurait ralenti le développement de ce TP.

3. **`net/http` de la stdlib suffisant** : Go 1.22 ajoute le routage avec paramètres de chemin (`GET /v1/checks/{id}`) directement dans `http.NewServeMux()`, sans dépendance externe. Cela réduit la surface d'attaque et simplifie la maintenance.

**Limite ressentie** : l'absence de génériques complets dans certaines abstractions (ex. typage des résultats de channel) oblige parfois à dupliquer du code ou à passer par `any`. Les génériques Go 1.18+ aident, mais leur ergonomie reste en dessous de Rust ou d'un langage fonctionnel typé.
