#!/bin/bash

RESULT=$(git log --oneline --decorate v0.2.0..v0.3.0)
echo $RESULT