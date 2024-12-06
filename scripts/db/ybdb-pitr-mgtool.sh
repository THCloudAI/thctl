#!/bin/bash

#######################################################################################
# Script Name: ybdb_pitr_mgtools.sh
# Description: YugabyteDB PITR Management Tools for backup and restore operations
# Author: OC
# Date Created: 2024-11-29
# Version: 1.0.1
#
# This script provides functionality to:
# - List all snapshot schedules
# - Delete specific snapshot schedules
# - Restore database to a specific point in time
#
# Requirements:
# - YugabyteDB installed and configured
# - Proper permissions to execute yb-admin commands
# - Write access to log directory
#######################################################################################

set -euo pipefail
export LANG=en_US.UTF-8
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Default configuration file location
readonly CONFIG_FILE="${CONFIG_FILE:-/etc/ybdb/ybdb-backup.conf}"
readonly SCRIPT_NAME=$(basename "$0")
readonly SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

# Function to get formatted timestamp
get_timestamp() {
    date '+%Y-%m-%d %H:%M:%S'
}

# Function to display usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]
Options:
  -l, --list     List all snapshot schedules
  -d, --delete   Delete a snapshot schedule (requires -i <schedule_id>)
  -r, --restore  Restore to a point in time (requires -i <schedule_id> and time options)
  -i <id>        Specify the schedule ID
  -c <file>      Specify config file (default: /etc/ybdb/ybdb-backup.conf)
  -h, --help     Display this help message

Time options for restore:
  -t <timestamp> Unix timestamp in microseconds
  -m <minutes>   Minutes ago (e.g., 5m)
  -H <hours>     Hours ago (e.g., 1h)
  -y <time>      YCQL timestamp (e.g., "2024-11-27 13:00-0700")

Examples:
  $0 -l                          # List all snapshot schedules
  $0 -d -i <schedule_id>         # Delete schedule
  $0 -r -i <id> -t 1651435200   # Restore using Unix timestamp
  $0 -r -i <id> -m 5            # Restore to 5 minutes ago
  $0 -r -i <id> -H 1            # Restore to 1 hour ago
EOF
    exit 1
}

# Function to load configuration
load_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "ERROR: Configuration file not found: $CONFIG_FILE"
        exit 1
    fi

    # Load configuration file
    # shellcheck source=/dev/null
    . "$CONFIG_FILE"

    # Validate required parameters
    local required_params=(LOG_DIR YB_ADMIN_PATH MASTER_ADDRESSES LOG_FILE_PITR_MGTOOLS)
    for param in "${required_params[@]}"; do
        if [ -z "${!param}" ]; then
            echo "ERROR: $param is not set in configuration file"
            exit 1
        fi
    done

    # Ensure log directory exists
    if [ ! -d "$LOG_DIR" ]; then
        if ! mkdir -p "$LOG_DIR"; then
            echo "ERROR: Failed to create log directory: $LOG_DIR"
            exit 1
        fi
    fi

    # Set log file path
    LOG_FILE="${LOG_DIR}/${LOG_FILE_PITR_MGTOOLS}"

    # Initialize log file
    write_log "=== YugabyteDB PITR Management Tool Started ===" "INFO" "no"
    write_log "Configuration loaded from: $CONFIG_FILE" "INFO" "no"
}

# Function for writing logs
write_log() {
    local message="$1"
    local log_level="${2:-INFO}"
    local display="${3:-no}"
    local timestamp

    timestamp=$(get_timestamp)

    if [ ! -d "$LOG_DIR" ]; then
        if ! mkdir -p "$LOG_DIR"; then
            echo "ERROR: Failed to create log directory: $LOG_DIR"
            exit 1
        fi
    fi

    printf "[%s] [%s] %s\n" "$timestamp" "$log_level" "$message" >> "$LOG_FILE"

    if [ "$display" = "yes" ] || [ "$log_level" = "ERROR" ]; then
        printf "[%s] %s: %s\n" "$timestamp" "$log_level" "$message" >&2
    fi
}

# Function to handle errors
handle_error() {
    local error_message="$1"
    write_log "$error_message" "ERROR" "yes"
    exit 1
}

# Function to check if schedule exists
check_schedule_exists() {
    local uuid="$1"
    local result

    write_log "Checking if schedule exists: $uuid" "INFO" "no"
    result=$(${YB_ADMIN_PATH}/yb-admin -master_addresses "$MASTER_ADDRESSES" list_snapshot_schedules 2>&1 | grep -w "$uuid" || true)

    if [ -n "$result" ]; then
        write_log "Schedule found: $uuid" "INFO" "no"
        return 0
    else
        write_log "Schedule not found: $uuid" "INFO" "no"
        return 1
    fi
}

# Function to perform delete schedule
delete_schedule() {
    local uuid="$1"
    if [ -z "$uuid" ]; then
        handle_error "Schedule ID is required for deletion"
    fi

    write_log "Starting delete operation for schedule ID: $uuid" "INFO" "no"

    if ! check_schedule_exists "$uuid"; then
        handle_error "Schedule ID '$uuid' does not exist"
    fi

    local result
    result=$(${YB_ADMIN_PATH}/yb-admin -master_addresses "$MASTER_ADDRESSES" delete_snapshot_schedule "$uuid" 2>&1)
    if [ $? -ne 0 ]; then
        handle_error "Delete operation failed: $result"
    fi

    write_log "Successfully deleted schedule with ID: $uuid" "INFO" "no"
    echo "Schedule deleted successfully."
}

# Function to perform list schedule
list_schedule() {
    write_log "Starting list operation" "INFO" "no"
    local result

    result=$(${YB_ADMIN_PATH}/yb-admin -master_addresses "$MASTER_ADDRESSES" list_snapshot_schedules 2>&1)
    if [ $? -ne 0 ]; then
        handle_error "Failed to list schedules: $result"
    fi

    if [ -z "$result" ]; then
        echo "No snapshot schedules found."
        write_log "No snapshot schedules found" "INFO" "no"
        return 0
    fi

    write_log "Schedule list retrieved successfully" "INFO" "no"
    write_log "Schedule list: $result" "INFO" "no"
    echo "$result"
}

# Function to perform restore schedule
perform_restore() {
    local schedule_id="$1"
    local restore_type="$2"
    local restore_value="$3"

    write_log "Starting restore operation" "INFO" "no"
    write_log "Schedule ID: $schedule_id" "INFO" "no"
    write_log "Restore type: $restore_type" "INFO" "no"
    write_log "Restore value: $restore_value" "INFO" "no"

    if [ -z "$schedule_id" ] || [ -z "$restore_type" ] || [ -z "$restore_value" ]; then
        handle_error "Missing required parameters for restore"
    fi

    if ! check_schedule_exists "$schedule_id"; then
        handle_error "Schedule ID '$schedule_id' does not exist"
    fi

    local restore_cmd
    case "$restore_type" in
        "timestamp")
            write_log "Restore type: Unix timestamp - $restore_value" "INFO" "no"
            restore_cmd="${YB_ADMIN_PATH}/yb-admin -master_addresses \"$MASTER_ADDRESSES\" restore_snapshot_schedule \"$schedule_id\" \"$restore_value\""
            ;;
        "minutes")
            write_log "Restore type: $restore_value minutes ago" "INFO" "no"
            restore_cmd="${YB_ADMIN_PATH}/yb-admin -master_addresses \"$MASTER_ADDRESSES\" restore_snapshot_schedule \"$schedule_id\" minus \"${restore_value}m\""
            ;;
        "hours")
            write_log "Restore type: $restore_value hours ago" "INFO" "no"
            restore_cmd="${YB_ADMIN_PATH}/yb-admin -master_addresses \"$MASTER_ADDRESSES\" restore_snapshot_schedule \"$schedule_id\" minus \"${restore_value}h\""
            ;;
        "ycql")
            write_log "Restore type: YCQL timestamp - $restore_value" "INFO" "no"
            restore_cmd="${YB_ADMIN_PATH}/yb-admin -master_addresses \"$MASTER_ADDRESSES\" restore_snapshot_schedule \"$schedule_id\" \"$restore_value\""
            ;;
        *)
            handle_error "Invalid restore type: $restore_type"
            ;;
    esac

    write_log "Executing restore command: $restore_cmd" "INFO" "no"
    local result
    result=$(eval "$restore_cmd" 2>&1)
    if [ $? -ne 0 ]; then
        handle_error "Restore operation failed: $result"
    fi

    write_log "Restore command output: $result" "INFO" "no"
    write_log "Restore operation completed successfully" "INFO" "no"
    echo "Restore completed successfully."
}

# Function to validate environment
validate_environment() {
    write_log "Validating environment" "INFO" "no"

    # Check if yb-admin exists and is executable
    if [ ! -x "${YB_ADMIN_PATH}/yb-admin" ]; then
        handle_error "yb-admin not found or not executable at ${YB_ADMIN_PATH}/yb-admin"
    fi

    # Check if we can connect to master
    local check_result
    check_result=$(${YB_ADMIN_PATH}/yb-admin -master_addresses "$MASTER_ADDRESSES" list_all_masters 2>&1)
    if [ $? -ne 0 ]; then
        handle_error "Cannot connect to YugabyteDB master at $MASTER_ADDRESSES: $check_result"
    fi

    write_log "Environment validation completed successfully" "INFO" "no"
}

# Function to cleanup on exit
cleanup() {
    local exit_code=$?
    write_log "Script execution completed with exit code: $exit_code" "INFO" "no"
    exit $exit_code
}

# Main function
main() {
    local SCHEDULE_ID=""
    local OPERATION=""
    local RESTORE_TYPE=""
    local RESTORE_VALUE=""

    # Set trap for cleanup
    trap cleanup EXIT

    # Parse command line options
    while [ $# -gt 0 ]; do
        case "$1" in
            -h|--help)
                show_usage
                ;;
            -c)
                CONFIG_FILE="$2"
                shift 2
                ;;
            -l|--list)
                OPERATION="list"
                shift
                ;;
            -d|--delete)
                OPERATION="delete"
                shift
                ;;
            -r|--restore)
                OPERATION="restore"
                shift
                ;;
            -i)
                SCHEDULE_ID="$2"
                shift 2
                ;;
            -t)
                RESTORE_TYPE="timestamp"
                RESTORE_VALUE="$2"
                shift 2
                ;;
            -m)
                RESTORE_TYPE="minutes"
                RESTORE_VALUE="$2"
                shift 2
                ;;
            -H)
                RESTORE_TYPE="hours"
                RESTORE_VALUE="$2"
                shift 2
                ;;
            -y)
                RESTORE_TYPE="ycql"
                RESTORE_VALUE="$2"
                shift 2
                ;;
            *)
                echo "ERROR: Unknown option: $1"
                show_usage
                ;;
        esac
    done

    # Load configuration and validate environment
    load_config
    validate_environment

    # Execute requested operation
    case "$OPERATION" in
        list)
            list_schedule
            ;;
        delete)
            if [ -z "$SCHEDULE_ID" ]; then
                echo "ERROR: Delete operation requires schedule ID (-i option)"
                show_usage
            fi
            delete_schedule "$SCHEDULE_ID"
            ;;
        restore)
            if [ -z "$SCHEDULE_ID" ]; then
                echo "ERROR: Restore operation requires schedule ID (-i option)"
                show_usage
            fi
            if [ -z "$RESTORE_TYPE" ] || [ -z "$RESTORE_VALUE" ]; then
                echo "ERROR: Restore operation requires time option (-t, -m, -H, or -y)"
                show_usage
            fi
            perform_restore "$SCHEDULE_ID" "$RESTORE_TYPE" "$RESTORE_VALUE"
            ;;
        *)
            echo "ERROR: No operation specified (-l, -d, or -r required)"
            show_usage
            ;;
    esac
}

# Script entry point
main "$@"