# Microservices

This directory contains all microservices organized by business domain:

- `fil/`: Filecoin related services
  - `sectors/`: Sector management services
    - `penalty/`: Sector penalty calculation service
  - `miners/`: Miner management services
  - `market/`: Market related services
- `eth/`: Ethereum related services
  - `tokens/`: Token management services
  - `contracts/`: Smart contract services
- `other/`: Services for other business domains

Each service is designed to be independently deployable and maintainable.
