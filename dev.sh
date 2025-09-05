#!/bin/bash

if [ ! "$(docker ps -q -f name=qoinly-tbdb-container)" ]; then
    echo "🚀 Starting development container..."
    docker-compose up -d
fi

echo "🎉 Ready! Opening Helix..."
docker exec -it qoinly-tbdb-container hx "$@"