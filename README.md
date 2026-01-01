# ShutterTrace

PRNU-based camera fingerprint experiments in Go.

Commands:
prnu enroll --camera <id> --in <folder> [--out <db>]
prnu verify --camera <id> --db <db> --img <path>

Current status:
Repo skeleton + CLI + meta storage.
Next steps:

-   image loading
-   denoise + residual
-   fingerprint estimation
-   correlation + PCE

Goal:
Understand PRNU limits while learning Go systems-style code.
