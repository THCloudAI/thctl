#!/bin/bash

#######################################################################################
# Script: ybdb-backup.sh
# Description: Backup script for YugabyteDB using YSQL interface
# Version: 1.0
# Date: 2024-11-27
# Author: OC
#######################################################################################

# Configuration file path
CONFIG_FILE="/etc/ybdb/ybdb-backup.conf"

# Function for writing logs
write_log() {
    local message="$1"
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S")
    echo "$timestamp - $message" >> "$LOG_FILE"
    echo "$timestamp - $message"
}

# Function for handling errors
handle_error() {
    local error_message="$1"
    write_log "ERROR: $error_message"
    [ -d "$TEMP_BACKUP_DIR" ] && rm -rf "$TEMP_BACKUP_DIR"
    exit 1
}

# Function to load configuration
load_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "ERROR: Configuration file not found: $CONFIG_FILE"
        exit 1
    fi

    # Source the configuration file
    source "$CONFIG_FILE"

    # Validate required configuration parameters
    local required_params=("BACKUP_DIR" "LOG_DIR" "TMP_DIR" "YSQL_HOST" "LOG_FILE_YSQL"
                         "BACKUP_PREFIX" "BACKUP_YSQL_SUFFIX" "YSQL_HOME")
    local missing_params=()

    for param in "${required_params[@]}"; do
        if [ -z "${!param}" ]; then
            missing_params+=("$param")
        fi
    done

    if [ ${#missing_params[@]} -ne 0 ]; then
        echo "ERROR: Missing required configuration parameters:"
        printf '%s\n' "${missing_params[@]}"
        exit 1
    fi

    # Create LOG_FILE variable after loading config
    LOG_FILE="${LOG_DIR}/${LOG_FILE_YSQL}"
}

# Function to check prerequisites
check_prerequisites() {
    write_log "Checking prerequisites..."

    # Check if YSQL_HOME is valid
    if [ -z "$YSQL_HOME" ]; then
        handle_error "YSQL_HOME is not set"
    fi
    write_log "YSQL_HOME is set to: $YSQL_HOME"

    # Check if ysql_dumpall exists and is executable
    YSQL_DUMP="${YSQL_HOME}/ysql_dumpall"
    if [ ! -x "$YSQL_DUMP" ]; then
        handle_error "ysql_dumpall not found or not executable at $YSQL_DUMP"
    fi
    write_log "ysql_dumpall found at: $YSQL_DUMP"

    # Simple database connection test
    if ! "${YSQL_HOME}/ysqlsh" -h "$YSQL_HOST" -c "\conninfo" >/dev/null 2>&1; then
        handle_error "Cannot connect to database at host: $YSQL_HOST"
    fi
    write_log "Database connection test successful"
}

# Function to create necessary directories
create_directories() {
    local dirs=("$BACKUP_DIR" "$LOG_DIR" "$TMP_DIR")
    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir" || handle_error "Failed to create directory: $dir"
            write_log "Created directory: $dir"
        fi
    done
}

# Function to perform backup
perform_backup() {
    # Create timestamp for backup
    local backup_timestamp=$(date +%Y%m%d-%H%M)
    local backup_dirname="${backup_timestamp}-ysql-fullbackup"
    TEMP_BACKUP_DIR="${TMP_DIR}/${backup_dirname}"

    # Create temporary backup directory
    mkdir -p "$TEMP_BACKUP_DIR" || handle_error "Failed to create temporary backup directory"
    write_log "Created temporary backup directory: $TEMP_BACKUP_DIR"

    # Set backup file paths
    local backup_sql="${TEMP_BACKUP_DIR}/backup.sql"
    local final_backup_file="${BACKUP_DIR}/${backup_dirname}.tar.gz"

    # Perform the backup
    write_log "Starting YSQL backup using ysql_dumpall..."
    if ! "$YSQL_DUMP" -h "$YSQL_HOST" > "$backup_sql" 2>> "$LOG_FILE"; then
        handle_error "Database backup failed"
    fi

    # Verify backup file exists and is not empty
    if [ ! -s "$backup_sql" ]; then
        handle_error "Backup file is empty or does not exist"
    fi
    write_log "Database backup completed successfully"

    # Create restore script
    create_restore_script "$backup_dirname"

    # Create tar.gz archive
    write_log "Creating backup archive..."
    cd "$TMP_DIR" || handle_error "Failed to change to temporary directory"
    if ! tar -czf "$final_backup_file" "$backup_dirname"; then
        handle_error "Failed to create backup archive"
    fi

    # Verify the archive was created successfully
    if [ ! -f "$final_backup_file" ]; then
        handle_error "Backup archive was not created successfully"
    fi

    write_log "Backup archive created successfully: $final_backup_file"

    # Cleanup temporary directory
    rm -rf "$TEMP_BACKUP_DIR"
    write_log "Cleaned up temporary backup directory"
}

# Function to create restore script
create_restore_script() {
    local backup_dirname="$1"
    local restore_script="${TEMP_BACKUP_DIR}/restore.sh"

    cat > "$restore_script" << EOF
#!/bin/bash
# Restore script for YSQL backup
# Backup name: ${backup_dirname}
# Created: $(date)

# Configuration file path
CONFIG_FILE="/etc/ybdb/ybdb-backup.conf"

# Load configuration
if [ ! -f "\$CONFIG_FILE" ]; then
    echo "ERROR: Configuration file not found: \$CONFIG_FILE"
    exit 1
fi

# Source the configuration file
source "\$CONFIG_FILE"

# Validate YSQL_HOME and YSQL_HOST
if [ -z "\$YSQL_HOME" ]; then
    echo "ERROR: YSQL_HOME is not set in configuration file"
    exit 1
fi

if [ -z "\$YSQL_HOST" ]; then
    echo "ERROR: YSQL_HOST is not set in configuration file"
    exit 1
fi

if [ -z "\$1" ]; then
    echo "Error: Please provide the path to the ${backup_dirname}.tar.gz file"
    exit 1
fi

BACKUP_FILE="\$1"
TEMP_DIR="/tmp/ybdb_restore_\$\$"

echo "Starting restore process..."
echo "Using backup file: \$BACKUP_FILE"
echo "Using YSQL_HOME: \$YSQL_HOME"
echo "Using YSQL_HOST: \$YSQL_HOST"

# Create temporary directory
mkdir -p "\$TEMP_DIR"

# Extract the backup
if ! tar -xzf "\$BACKUP_FILE" -C "\$TEMP_DIR"; then
    echo "Failed to extract backup file"
    rm -rf "\$TEMP_DIR"
    exit 1
fi

# Check if ysqlsh exists and is executable
if [ ! -x "\$YSQL_HOME/ysqlsh" ]; then
    echo "ERROR: ysqlsh not found or not executable at \$YSQL_HOME/ysqlsh"
    rm -rf "\$TEMP_DIR"
    exit 1
fi

# Restore the database
if "\$YSQL_HOME/ysqlsh" -h "\$YSQL_HOST" -f "\$TEMP_DIR/${backup_dirname}/backup.sql"; then
    echo "Database restore completed successfully"
else
    echo "Database restore failed"
    rm -rf "\$TEMP_DIR"
    exit 1
fi

# Cleanup
rm -rf "\$TEMP_DIR"
echo "Restore process completed"
EOF

    chmod +x "$restore_script"
    write_log "Created restore script: $restore_script"
}

# Function to cleanup old backups (if needed)
cleanup_old_backups() {
    # Add cleanup logic here if needed
    write_log "Cleanup process completed"
}

# Main function
main() {
    # Load configuration
    load_config

    # Initialize backup process
    write_log "=== Starting YSQL backup process ==="

    # Run prerequisite checks
    check_prerequisites

    # Create necessary directories
    create_directories

    # Perform the backup
    perform_backup

    # Cleanup old backups if needed
    cleanup_old_backups

    write_log "=== Backup process completed ==="
}

# Execute main function
main