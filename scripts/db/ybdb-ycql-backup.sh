#!/bin/bash

#######################################################################################
# Script: ybdb_ycql_backup.sh
# Description: Enhanced full backup script for YugabyteDB using YCQL interface
# Version: 2.6
# Date: 2024-11-27
# Author: OC
#######################################################################################

# Load configuration
CONFIG_FILE="/etc/ybdb/ybdb-backup.conf"

# Function to read configuration
load_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        handle_error "Configuration file not found: $CONFIG_FILE"
    fi

    source "$CONFIG_FILE"

    # Validate required configuration parameters
    local required_params=("BACKUP_DIR" "LOG_DIR" "TMP_DIR" "DB_HOST" "LOG_FILE_YCQL"
                         "BACKUP_PREFIX" "BACKUP_YCQL_SUFFIX" "CHUNK_SIZE" "INGEST_RATE")
    local missing_params=()

    for param in "${required_params[@]}"; do
        if [ -z "${!param}" ]; then
            missing_params+=("$param")
        fi
    done

    if [ ${#missing_params[@]} -ne 0 ]; then
        handle_error "Missing required configuration parameters: ${missing_params[*]}"
    fi

    LOG_FILE="${LOG_DIR}/${LOG_FILE_YCQL}"
}

# Function to write logs with timestamp (file only)
write_log() {
    local log_time=$(date "+%Y-%m-%d %H:%M:%S")
    echo "[$log_time] $1" >> "$LOG_FILE"
}

# Function to handle errors
handle_error() {
    echo "ERROR: $1" >&2
    write_log "ERROR: $1"
    write_log "Backup process failed"
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    if [ -z "$YCQL_HOME" ]; then
        handle_error "YCQL_HOME environment variable is not set. Please set it to YugabyteDB installation directory."
    fi
    write_log "YCQL_HOME is set to: $YCQL_HOME"

    YCQL_SHELL="${YCQL_HOME}/ycqlsh"
    if [ ! -x "$YCQL_SHELL" ]; then
        handle_error "ycqlsh not found or not executable at $YCQL_SHELL"
    fi
    write_log "ycqlsh found at: $YCQL_SHELL"
}

# Function to check directories and disk space
check_directories() {
    # Check available disk space (minimum 1GB required)
    local available_space=$(df -BG "$BACKUP_DIR" | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$available_space" -lt 1 ]; then
        handle_error "Insufficient disk space in backup directory. Available: ${available_space}GB, Required: 1GB"
    fi

    for dir in "$LOG_DIR" "$TMP_DIR" "$BACKUP_DIR"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir" || handle_error "Failed to create directory: $dir"
            write_log "Created directory: $dir"
        fi

        if [ ! -w "$dir" ]; then
            handle_error "Directory not writable: $dir"
        fi
    done
}

# Function to get all tables in a keyspace

get_tables() {
    local keyspace=$1
    local tables_file="${DATA_DIR}/${keyspace}/tables.txt"

    write_log "Getting tables for keyspace: $keyspace"

    # Get only table names from the keyspace using a more precise query
    ${YCQL_SHELL} ${DB_HOST} -e "SELECT table_name FROM system_schema.tables WHERE keyspace_name = '${keyspace}';" | \
    awk 'NR>3 {
        # Skip headers and separator lines
        if ($0 !~ /^-+$/ && $0 !~ /^[[:space:]]*$/ && $1 !~ /^table_name$/) {
            # Remove leading/trailing whitespace and (x rows) pattern
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", $0)
            gsub(/\([0-9]+ rows\)/, "", $0)
            # Only output non-empty lines
            if (length($0) > 0) print $0
        }
    }' > "$tables_file"

    if [ $? -eq 0 ] && [ -s "$tables_file" ]; then
        write_log "Successfully retrieved tables list for keyspace ${keyspace}:"
        cat "$tables_file" | while read -r table; do
            write_log "  - $table"
            # Export data for each table
            export_table_data "$keyspace" "$table"
        done
        return 0
    else
        write_log "ERROR: Failed to get tables list or no tables found for keyspace ${keyspace}"
        return 1
    fi
}


export_table_data() {
    local keyspace=$1
    local table=$2
    local output_file="${DATA_DIR}/${keyspace}/${table}.csv"

    write_log "Exporting data from ${keyspace}.${table} to ${output_file}"

    # Get column names from system schema
    local columns=$(${YCQL_SHELL} ${DB_HOST} -e "
        SELECT column_name
        FROM system_schema.columns
        WHERE keyspace_name='${keyspace}'
        AND table_name='${table}'
        ORDER BY position;" | \
        awk 'NR>3 {if ($0 !~ /^-+$/ && $0 !~ /^\(.*rows\)$/ && length($0) > 0) print $0}' | \
        tr '\n' ',' | sed 's/,$//')

    # Generate column names for SELECT statement
    local select_columns=$(echo $columns | sed 's/,/","/g')
    select_columns="\"$select_columns\""

    # Export data in CSV format with specific columns
    ${YCQL_SHELL} ${DB_HOST} -e "
        COPY ${keyspace}.${table} (${columns})
        TO '${output_file}'
        WITH HEADER = TRUE
        AND DELIMITER = ','
        AND QUOTE = '\"'
        AND ESCAPE = '\"'
        AND NULL = '';"

    if [ $? -eq 0 ] && [ -s "$output_file" ]; then
        write_log "Successfully exported data from ${keyspace}.${table}"
        # Verify CSV format by logging first two lines
        head -n 2 "$output_file" >> "${LOG_FILE}"
        return 0
    else
        write_log "ERROR: Failed to export data from ${keyspace}.${table}"
        return 1
    fi
}

# Helper function to verify table structure before export
verify_table_structure() {
    local keyspace=$1
    local table=$2

    write_log "Verifying table structure for ${keyspace}.${table}"
    ${YCQL_SHELL} ${DB_HOST} -e "
        SELECT column_name, type
        FROM system_schema.columns
        WHERE keyspace_name='${keyspace}'
        AND table_name='${table}'
        ORDER BY position;"
}


# Simplified schema export function using DESC command
export_schema() {
    local keyspace=$1
    local schema_file="${SCHEMAS_DIR}/${keyspace}.cql"
    write_log "Exporting schema for keyspace: $keyspace"

    # Export entire keyspace schema using DESC command
    ${YCQL_SHELL} ${DB_HOST} -e "DESC KEYSPACE $keyspace" > "$schema_file" 2>/dev/null

    if [ $? -ne 0 ]; then
        handle_error "Failed to export schema for keyspace $keyspace"
    fi

    write_log "Schema export completed for keyspace: $keyspace"
}

# Simplified table data export function using COPY TO command
export_table_data() {
    local keyspace=$1
    local table=$2
    local csv_file="${DATA_DIR}/${keyspace}/${table}.csv"
    local error_log="${DATA_DIR}/${keyspace}/${table}_error.log"

    write_log "Starting export for $keyspace.$table"

    # 使用管道方式导出数据
    ${YCQL_SHELL} ${DB_HOST} -e "SELECT * FROM $keyspace.$table;" | \
    awk 'BEGIN {FS="|"; OFS=","}
         NR==1 {next} # 跳过第一行空行
         NR==2 {gsub(/ /, "", $0); print} # 处理表头
         NR>3 {gsub(/^ +| +$/, "", $0); if($0 !~ /^-+$/) print}' > "$csv_file" 2>>"$error_log"

    # 验证导出结果
    if [ $? -eq 0 ] && [ -s "$csv_file" ]; then
        local row_count=$(wc -l < "$csv_file")
        write_log "Export completed successfully. Total rows (including header): $row_count"

        # 显示样本数据
        write_log "Sample of exported data (first 2 lines):"
        head -n 2 "$csv_file" | while read -r line; do
            write_log "  $line"
        done

        # 显示文件大小
        local file_size=$(ls -lh "$csv_file" | awk '{print $5}')
        write_log "CSV file size: $file_size"

        return 0
    else
        write_log "ERROR: Export failed or produced empty file"
        if [ -f "$error_log" ]; then
            write_log "Error details:"
            cat "$error_log"
        fi
        return 1
    fi
}



# Function to create restore script
create_restore_script() {
    local restore_path=$1
    cat > "$restore_path" << 'EOF'
#!/bin/bash
# YugabyteDB YCQL Restore Script
# Usage: ./restore.sh -h <host> [--ycqlsh path_to_ycqlsh]

# Default values
YCQL_SHELL="ycqlsh"
HOST=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            HOST="$2"
            shift 2
            ;;
        --ycqlsh)
            YCQL_SHELL="$2"
            shift 2
            ;;
        *)
            echo "Unknown parameter: $1" >&2
            echo "Usage: ./restore.sh -h <host> [--ycqlsh path_to_ycqlsh]" >&2
            exit 1
            ;;
    esac
done

# Verify host parameter
if [ -z "$HOST" ]; then
    echo "Error: Host parameter (-h) is required" >&2
    echo "Usage: ./restore.sh -h <host> [--ycqlsh path_to_ycqlsh]" >&2
    exit 1
fi

echo "Starting restore process..."

# Step 1: Restore schemas
echo "Restoring schemas..."
for schema_file in schemas/*.cql; do
    if [ -f "$schema_file" ]; then
        echo "Processing schema: $schema_file"
        $YCQL_SHELL $HOST -f "$schema_file"
        if [ $? -ne 0 ]; then
            echo "Error restoring schema from $schema_file" >&2
            exit 1
        fi
    fi
done

# Step 2: Restore data from CSV files
echo "Restoring data..."
for keyspace_dir in data/*; do
    if [ -d "$keyspace_dir" ]; then
        keyspace=$(basename "$keyspace_dir")
        echo "Processing keyspace: $keyspace"

        for csv_file in "$keyspace_dir"/*.csv; do
            if [ -f "$csv_file" ]; then
                table=$(basename "$csv_file" .csv)
                echo "Restoring data for $keyspace.$table"

                # Execute COPY FROM command using ycqlsh
                $YCQL_SHELL $HOST -e "COPY $keyspace.$table FROM '$csv_file' WITH HEADER = true;"

                if [ $? -ne 0 ]; then
                    echo "Error restoring data for $keyspace.$table from $csv_file" >&2
                    exit 1
                fi

                echo "Successfully restored data for $keyspace.$table"
            fi
        done
    fi
done

echo "Restore completed successfully!"
EOF
    chmod +x "$restore_path"
    write_log "Created restore script: $restore_path"
}

# Enhanced backup function
perform_backup() {
    DATE_FORMAT=$(date +%Y%m%d-%H%M)
    BACKUP_NAME="${DATE_FORMAT}-${BACKUP_PREFIX}-${BACKUP_YCQL_SUFFIX}"
    TMP_BACKUP_DIR="${TMP_DIR}/${BACKUP_NAME}"
    SCHEMAS_DIR="${TMP_BACKUP_DIR}/schemas"
    DATA_DIR="${TMP_BACKUP_DIR}/data"
    FINAL_BACKUP="${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
    local all_tables_exported=true  # Flag to track tables backup status

    # Create temporary backup directory structure
    mkdir -p "${SCHEMAS_DIR}" "${DATA_DIR}" || handle_error "Failed to create temporary backup directory"
    write_log "Created temporary backup directory: ${TMP_BACKUP_DIR}"

    # Create restore script
    write_log "Creating restore script"
    create_restore_script "${TMP_BACKUP_DIR}/restore.sh"

    # Check if keyspaces variable is set in the config file
    if [ -z "$keyspaces" ]; then
        handle_error "Keyspaces variable is not set in the configuration file."
    fi

    # Process each keyspace defined in the keyspaces variable
    IFS=',' read -r -a keyspace_array <<< "$keyspaces"
    for keyspace in "${keyspace_array[@]}"; do
        keyspace=$(echo "$keyspace" | xargs)  # Trim whitespace
        write_log "Processing keyspace: $keyspace"
        mkdir -p "${DATA_DIR}/${keyspace}"

        # Export schema
        export_schema "$keyspace"

        # Export table data
        local tables=$(get_tables "$keyspace")
        if [ $? -eq 0 ]; then
            echo "$tables" | while read -r table; do
                if [ ! -z "$table" ]; then  # Only process non-empty table names
                    export_table_data "$keyspace" "$table" || all_tables_exported=false
                fi
            done
        fi
    done

    # Check if all tables were exported successfully
    if [ "$all_tables_exported" != true ]; then
        handle_error "Some tables failed to export. Exiting backup process."
    fi

    # Create compressed backup
    write_log "Creating compressed backup"
    cd "${TMP_DIR}" && tar -czf "${BACKUP_NAME}.tar.gz" "$(basename ${TMP_BACKUP_DIR})" || handle_error "Backup compression failed"

    # Calculate MD5 checksum
    md5sum "${BACKUP_NAME}.tar.gz" > "${BACKUP_NAME}.tar.gz.md5" || handle_error "Failed to create MD5 checksum"

    # Move backup files to final location
    mv "${BACKUP_NAME}.tar.gz" "${FINAL_BACKUP}" || handle_error "Failed to move backup to final location"
    mv "${BACKUP_NAME}.tar.gz.md5" "${FINAL_BACKUP}.md5" || handle_error "Failed to move MD5 file"

    # Clean up
    rm -rf "${TMP_BACKUP_DIR}"
    write_log "Temporary files cleaned up"

    # Verify final backup
    if [ ! -s "${FINAL_BACKUP}" ]; then
        handle_error "Backup file is empty or was not created"
    fi

    write_log "Backup completed successfully: ${FINAL_BACKUP}"
    write_log "MD5 checksum: $(cat ${FINAL_BACKUP}.md5)"
}

# Main execution
main() {
    local backup_success=false

    # Load configuration
    load_config

    write_log "=== Starting YCQL backup process ==="
    write_log "Backup started at: $(date)"

    # Use trap to catch errors
    trap 'handle_error "An unexpected error occurred."' ERR

    # Check prerequisites
    check_prerequisites

    # Check directories
    check_directories

    # Perform backup
    perform_backup && backup_success=true

    # Output different conclusion information based on backup result
    if [ "$backup_success" = true ]; then
        write_log "Backup completed at: $(date)"
        write_log "=== Backup process completed successfully ==="
    else
        write_log "Backup failed at: $(date)"
        write_log "=== Backup process failed ==="
    fi
}

# Execute main function
main "$@"