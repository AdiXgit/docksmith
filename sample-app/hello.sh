#!/bin/sh
echo "Message is: $MESSAGE"
echo "Sample file contents:"
cat data.txt
echo "Writing inside container root..."
echo "inside-container" > /tmp/container-note.txt
