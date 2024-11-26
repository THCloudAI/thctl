#!/bin/bash

# Log settings
LOG_DIR="/root/var/logs"
LOG_FILE="${LOG_DIR}/thc-ybdb-snapshot.log"
UUID_FILE="${LOG_DIR}/thc-latest-snapshot-uuid"

# Create log directory if it doesn't exist
mkdir -p "${LOG_DIR}" || {
    echo "Error: Failed to create log directory ${LOG_DIR}"
    exit 1
}

# Ensure log files exist (but do not overwrite if they already exist)
if [ ! -f "${LOG_FILE}" ]; then
    touch "${LOG_FILE}" || {
        echo "Error: Failed to create log file ${LOG_FILE}"
        exit 1
    }
fi

if [ ! -f "${UUID_FILE}" ]; then
    touch "${UUID_FILE}" || {
        echo "Error: Failed to create UUID file ${UUID_FILE}"
        exit 1
    }
fi

# Logging function
log() {
    printf "[%s] %s\n" "$(date '+%Y-%m-%d %H:%M:%S')" "$1" >> "${LOG_FILE}"
}

# Save UUID function
save_uuid() {
    local uuid=$1
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Append the new UUID with timestamp
    printf "[%s] %s\n" "$timestamp" "$uuid" >> "$UUID_FILE"
}

# Set YugabyteDB master node IP addresses
MASTER_ADDRESSES="10.10.3.221,10.10.3.118,10.10.3.119"

# Set database name
DATABASE_NAME="yugabyte"

# Set the path to the yb-admin binary using whereis
YB_ADMIN_PATH=$(whereis yb-admin | awk '{print $2}')

# Check if yb-admin was found
if [ -z "$YB_ADMIN_PATH" ]; then
    log "Error: yb-admin not found in system PATH"
    exit 1
fi

# Log the path being used
log "Using yb-admin from: ${YB_ADMIN_PATH}"

# Create a snapshot for the database
log "Creating snapshot for database ${DATABASE_NAME}..."
SNAPSHOT_OUTPUT=$($YB_ADMIN_PATH -master_addresses $MASTER_ADDRESSES create_database_snapshot ysql.$DATABASE_NAME 2>&1)

# Check if snapshot creation was successful
if [[ $SNAPSHOT_OUTPUT == *"Started snapshot creation"* ]]; then
    # Extract just the UUID from the output
    CLEAN_UUID=$(echo "$SNAPSHOT_OUTPUT" | sed 's/Started snapshot creation: //')

    log "Snapshot creation successful"
    log "Snapshot UUID: ${CLEAN_UUID}"

    # Save UUID with timestamp to UUID file
    save_uuid "$CLEAN_UUID"
    log "Snapshot UUID appended to: ${UUID_FILE}"
else
    log "Error in snapshot creation: $SNAPSHOT_OUTPUT"
    exit 1
fi