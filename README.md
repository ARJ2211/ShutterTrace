# ShutterTrace

ShutterTrace is a camera forensics project built around PRNU (Photo Response Non-Uniformity).
The goal is simple: check if two images were likely taken from the same physical camera sensor.

This project focuses on correctness, reproducibility, and learning. Code is written step by step,
with clear separation between denoising, fingerprint extraction, and verification.

## What it does

-   Extracts PRNU noise from images
-   Builds a camera fingerprint from reference images
-   Verifies a test image against a camera database
-   Supports PCE and Pearson correlation metrics
-   Tile-based matching for robustness

## Why this project

Most PRNU explanations stay very high level. This repo tries to go lower.
You can read the code, run experiments, and actually see where things break.
It is meant for people who want to understand camera fingerprinting, not just use it.

## Tech

-   Go
-   Basic image processing
-   Classical signal processing (no ML)
-   CLI-first design

## Status

Work in progress.
Some parts are experimental and results can vary based on image content and resolution.
That is expected and part of the learning.

## Usage (example)

```bash
./ShutterTrace enroll --camera iphone_14 --img ./datasets/enroll/
./ShutterTrace verify --camera iphone_14 --img ./datasets/test/test.jpg --metric both
```

## Disclaimer

This is not a court-ready forensic tool.
It is a research and learning project.

If you are reading this and things feel slightly rough, that is intentional.
