#!/bin/bash

#######################################################################################
# Script: ybdb-pitr-schedule.sh
# Description: Schedule script for PITR backup of YugabyteDB
# Version: 1.0
# Date: 2024-11-29
# Author: OC
#######################################################################################

# Configuration file path
CONFIG_FILE="/etc/ybdb/ybdb-backup.conf"

# Function to load configuration
load_config() {
    if [ -f "$CONFIG_FILE" ]; then
        source "$CONFIG_FILE"
        # Create LOG_FILE path using LOG_DIR and LOG_FILE_PITR from config
        LOG_FILE="${LOG_DIR}/${LOG_FILE_PITR}"
    else
        echo "ERROR: Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
}

# Function for writing logs
write_log() {
    # Check if LOG_DIR exists, if not create it
    if [ ! -d "$LOG_DIR" ]; then
        mkdir -p "$LOG_DIR" || handle_error "Failed to create log directory: $LOG_DIR"
    fi

    local message="$1"
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S")
    echo "[$timestamp] $message" >> "$LOG_FILE"
    echo "[$timestamp] $message"
}

# Function to handle errors
handle_error() {
    local error_message="$1"
    write_log "ERROR: $error_message"
    exit 1
}

# Function to verify prerequisites
check_prerequisites() {
    # Verify required config parameters
    local required_params=("LOG_DIR" "LOG_FILE_PITR" "YB_ADMIN_PATH" "MASTER_ADDRESSES"
                         "PRODUCES_TIME" "RETAINS_TIME" "DATABASE_NAME")

    for param in "${required_params[@]}"; do
        if [ -z "${!param}" ]; then
            handle_error "Required parameter ${param} is not set in configuration file"
        fi
    done

    # Verify yb-admin exists
    if [ ! -x "${YB_ADMIN_PATH}/yb-admin" ]; then
        handle_error "yb-admin not found or not executable at ${YB_ADMIN_PATH}/yb-admin"
    fi
}

# Function to perform PITR backup
perform_pitr_backup() {
    write_log "Starting PITR schedule creation process..."

    # Create schedule for YSQL database
    write_log "Creating snapshot schedule for YSQL database: ${DATABASE_NAME}"
    YSQL_RESULT=$(${YB_ADMIN_PATH}/yb-admin -master_addresses $MASTER_ADDRESSES create_snapshot_schedule ${PRODUCES_TIME} ${RETAINS_TIME} ysql.${DATABASE_NAME} 2>&1)
    if [ $? -ne 0 ]; then
        if echo "$YSQL_RESULT" | grep -q "OBJECT_ALREADY_PRESENT"; then
            write_log "WARNING: YSQL snapshot schedule already exists for database ${DATABASE_NAME}"
            write_log "Detail: $YSQL_RESULT"
        else
            handle_error "Failed to create YSQL snapshot schedule: $YSQL_RESULT"
        fi
    else
        write_log "Successfully created YSQL snapshot schedule"
    fi

    # Create schedule for YCQL keyspace
    write_log "Creating snapshot schedule for YCQL keyspaces: ${keyspaces}"
    YCQL_RESULT=$(${YB_ADMIN_PATH}/yb-admin -master_addresses $MASTER_ADDRESSES create_snapshot_schedule ${PRODUCES_TIME} ${RETAINS_TIME} ${keyspaces} 2>&1)
    if [ $? -ne 0 ]; then
        if echo "$YCQL_RESULT" | grep -q "OBJECT_ALREADY_PRESENT"; then
            write_log "WARNING: YCQL snapshot schedule already exists for keyspaces ${keyspaces}"
            write_log "Detail: $YCQL_RESULT"
        else
            handle_error "Failed to create YCQL snapshot schedule: $YCQL_RESULT"
        fi
    else
        write_log "Successfully created YCQL snapshot schedule"
    fi

    write_log "PITR schedule creation completed successfully"
}

# Main function
main() {
    # Load configuration
    load_config

    # Check prerequisites
    check_prerequisites

    # Start log entry
    write_log "=== Starting PITR schedule creation ==="

    # Execute PITR backup
    perform_pitr_backup

    # End log entry
    write_log "=== PITR schedule creation completed ==="
}

# Execute main function
main