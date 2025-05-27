## 1.1.0 (2025-05-27)

### Feat

- add commitizen config and version output
- test faster start times
- switch directly to workspace
- add customization option to use aliases for docker registries
- add "localWorkspaceFolderBasename" variable
- add context option without testing it
- load a global config first
- add logging and debug switch and cleanup command

### Fix

- replace only the value currently active, not all matches
- docker only allows lower case names
- dont run postStartCommand when container is already running
- prefix is the correct term, not suffix
- create log file only when file logging is active
- do not crash on env variable not set
- catch empty image names during global clean
- use dockerfile directory for building and redirect build to stdout
