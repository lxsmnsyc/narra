#!/bin/bash
# Copies this case's starter package into the run's empty working folder.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
cp -R "$here/starter/." .
