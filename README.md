# devcli

´devcli´ ist ein Command Line Tool um Devcontainer zu verwalten und zu starten.

´devcli´ baut nach einer ´.devcontainer/devcontainer.json´ ein Docker Image und startet es.
Der Container wird dabei in einer Bash Endlosschleife gehalten, in die sich ´devcli´ dann
mehrfach per ´docker exec´ verbinden kann.

Der Ablauf ist dabei wie folgt:

- Hash aus Config berechnen
- Prüfen ob Image bereits existiert, wenn nein, dann baue oder pulle Image
- Prüfen ob ein Container bereits existiert, wenn nein, dann starte in bash Endlosschleife
- falls erster Start, dann führe PostCreateCommands aus
- führe per ´docker exec´ bash Shell in Container aus


´devcli´ kann auch die gebauten Images und Container wieder löschen, siehe dazu die ´--help´
Funktion.
