#!/usr/bin/env bash
set -euo pipefail
cd fixtures
python3 -m unittest tests/test_urls.py
