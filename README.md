# devcli

`devcli` ist ein Command Line Tool um Devcontainer zu verwalten und zu starten.

`devcli` baut nach einer `.devcontainer/devcontainer.json` ein Docker Image und startet es.
Der Container wird dabei in einer Bash Endlosschleife gehalten, in die sich `devcli` dann
mehrfach per `docker exec` verbinden kann.

Der Ablauf ist dabei wie folgt:

- Hash aus Config berechnen
- Prüfen ob Image bereits existiert, wenn nein, dann baue oder pulle Image
- Prüfen ob ein Container bereits existiert, wenn nein, dann starte in bash Endlosschleife
- falls erster Start, dann führe PostCreateCommands aus
- führe per `docker exec` bash Shell in Container aus

## Subcommands

`devcli` kann auch die gebauten Images und Container wieder löschen, siehe dazu die `--help`
Funktion.

## globale Config

`devcli` prüft ob eine globale Config unter `~/.config/devcli/.devcontainer/devcontainer.json`
vorhanden ist. Wenn ja, wird diese zuerst geladen und dann von der Projekt-Config überschrieben
oder erweitert.
Eindeutige Felder wie `Image` oder `Dockerfile` werden überschrieben, Listen, wie `Mounts` oder
`PostStartCommand` werden ergänzt.
So kann zum Beispiel eine globaler `PostCreateCommand` hinzugefügt werden, der `nvim` installiert
und eine Config von Github lädt:
```
{
    "postCreateCommand": "cd && mkdir -p .config && cd .config && git clone https://github.com/johndoe2991/nvim && cd && wget https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.tar.gz && sudo tar -xf nvim-linux-x86_64.tar.gz -C /usr --strip-components=1"
}
```
