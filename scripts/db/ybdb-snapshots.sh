#!/bin/bash

# Color configuration
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Display logo
cat << "EOF"
 _____ _   _  ____ _                 _      _    ___
|_   _| | | |/ ___| | ___  _   _  __| |    / \  |_ _|
  | | | |_| | |   | |/ _ \| | | |/ _` |   / _ \  | |
  | | |  _  | |___| | (_) | |_| | (_| |_ / ___ \ | |
  |_| |_| |_|\____|_|\___/ \__,_|\__,_(_)_/   \_\___|
EOF
echo "Accelerate AI intelligent storage and computing"
echo "------------------------------------------------"
echo ""

# Configuration file path
CONFIG_FILE="/var/logs/.config.cfg"

# Log settings (will be set after configuration)
LOG_DIR=""
LOG_FILE=""
UUID_FILE=""

# Logging function
log() {
    printf "[%s] %b\n" "$(date '+%Y-%m-%d %H:%M:%S')" "$1" | tee -a "${LOG_FILE}"
}

# Save UUID function
save_uuid() {
    local uuid=$1
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Check if the file exists and is non-empty
    if [ -f "$UUID_FILE" ] && [ -s "$UUID_FILE" ]; then
        # Ensure the file doesn't end with an incomplete entry
        [ -z "$(tail -c1 "$UUID_FILE")" ] || echo "" >> "$UUID_FILE"
    fi

    # Append the new UUID with timestamp
    printf "[%s] %s\n" "$timestamp" "$uuid" >> "$UUID_FILE"
}

# Function to validate and set configuration parameters
configure_script() {
    echo -e "${GREEN}Welcome to the YugabyteDB Snapshot Management Script Configuration${NC}"
    echo "Please provide the following parameters:"

    # Validate LOG_DIR
    while true; do
        read -p "Enter the log directory path (e.g., /var/logs): " LOG_DIR_INPUT
        if [[ -z "$LOG_DIR_INPUT" ]]; then
            echo -e "${RED}Log directory cannot be empty. Please enter a valid directory path.${NC}"
        elif [[ -d "$LOG_DIR_INPUT" || ! -e "$LOG_DIR_INPUT" ]]; then
            LOG_DIR="$LOG_DIR_INPUT"
            break
        else
            echo -e "${RED}Invalid directory path. Please enter a valid path or a new path to be created.${NC}"
        fi
    done

    # Validate MASTER_ADDRESSES
    while true; do
        read -p "Enter the master addresses (e.g., 192.168.1.1:7100,192.168.1.2:7100): " MASTER_ADDRESSES_INPUT
        if [[ "$MASTER_ADDRESSES_INPUT" =~ ^[0-9a-zA-Z.,:]+$ ]]; then
            MASTER_ADDRESSES="$MASTER_ADDRESSES_INPUT"
            break
        else
            echo -e "${RED}Invalid master addresses format. Please enter comma-separated IP:port pairs.${NC}"
        fi
    done

    # Validate DATABASE_NAME
    while true; do
        read -p "Enter the database name (e.g., yugabyte): " DATABASE_NAME_INPUT
        if [[ "$DATABASE_NAME_INPUT" =~ ^[a-zA-Z0-9_]+$ ]]; then
            DATABASE_NAME="$DATABASE_NAME_INPUT"
            break
        else
            echo -e "${RED}Invalid database name. Please use alphanumeric characters and underscores only.${NC}"
        fi
    done

    # Save configuration to file
    mkdir -p "$(dirname "$CONFIG_FILE")"
    cat > "$CONFIG_FILE" << EOF
LOG_DIR="$LOG_DIR"
MASTER_ADDRESSES="$MASTER_ADDRESSES"
DATABASE_NAME="$DATABASE_NAME"
EOF

    echo -e "${GREEN}Configuration saved successfully!${NC}"
}

# Load configuration if available
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    configure_script
fi

# Set log files after configuration
LOG_FILE="${LOG_DIR}/snapshot_ybdb_$(date '+%Y%m%d%H').log"
UUID_FILE="${LOG_DIR}/latest_snapshot_uuid"

# Create log directory if it doesn't exist
mkdir -p "${LOG_DIR}" || {
    echo -e "${RED}Error: Failed to create log directory ${LOG_DIR}${NC}"
    exit 1
}

# Set the path to the yb-admin binary using whereis
YB_ADMIN_PATH=$(whereis yb-admin | awk '{print $2}')

# Check if yb-admin was found
if [ -z "$YB_ADMIN_PATH" ]; then
    log "${RED}Error: yb-admin not found in system PATH${NC}"
    exit 1
fi

# Print the path being used (for verification)
log "Using yb-admin from: ${GREEN}${YB_ADMIN_PATH}${NC}"

# Function to create snapshot
create_snapshot() {
    log "Creating snapshot for database ${GREEN}${DATABASE_NAME}${NC}..."
    SNAPSHOT_UUID=$($YB_ADMIN_PATH -master_addresses $MASTER_ADDRESSES create_database_snapshot ysql.$DATABASE_NAME)

    # Check if snapshot creation was successful
    if [[ $SNAPSHOT_UUID == *"Started snapshot creation"* ]]; then
        # Extract just the UUID from the output
        CLEAN_UUID=$(echo "$SNAPSHOT_UUID" | sed 's/Started snapshot creation: //')

        log "Snapshot creation: ${GREEN}successful${NC}"
        log "Snapshot UUID: ${GREEN}${CLEAN_UUID}${NC}"

        # Save UUID with timestamp to UUID file (append mode)
        save_uuid "$CLEAN_UUID"
        log "Snapshot UUID appended to: ${GREEN}${UUID_FILE}${NC}"

        # Display last 5 entries from UUID file
        echo -e "\nLast 5 snapshot records:"
        tail -n 5 "$UUID_FILE"
    else
        log "${RED}Error in snapshot creation!${NC}"
        exit 1
    fi
}

# Function to restore snapshot
restore_snapshot() {
    local UUID_TO_RESTORE=$1

    log "Restoring snapshot for database ${GREEN}${DATABASE_NAME}${NC}..."

    # Check if UUID is provided as argument
    if [ -z "$UUID_TO_RESTORE" ]; then
        log "${RED}Error: UUID argument is missing for restore operation.${NC}"
        exit 1
    fi

    log "Restoring from snapshot UUID: ${GREEN}${UUID_TO_RESTORE}${NC}"

    # Call the restore snapshot command using the UUID
    RESTORE_OUTPUT=$($YB_ADMIN_PATH -master_addresses $MASTER_ADDRESSES restore_snapshot $UUID_TO_RESTORE 2>&1)

    if [[ $RESTORE_OUTPUT == *"Restored snapshot"* ]]; then
        log "Snapshot restored ${GREEN}successfully${NC}"
    else
        log "${RED}Error during snapshot restore: ${RESTORE_OUTPUT}${NC}"
        exit 1
    fi
}

# Function to delete snapshot
delete_snapshot() {
    local UUID_TO_DELETE=$1

    log "Deleting snapshot for database ${GREEN}${DATABASE_NAME}${NC}..."

    # Check if UUID is provided as argument
    if [ -z "$UUID_TO_DELETE" ]; then
        log "${RED}Error: UUID argument is missing for delete operation.${NC}"
        exit 1
    fi

    log "Deleting snapshot UUID: ${GREEN}${UUID_TO_DELETE}${NC}"

    # Call the delete snapshot command using the UUID
    DELETE_OUTPUT=$($YB_ADMIN_PATH -master_addresses $MASTER_ADDRESSES delete_snapshot $UUID_TO_DELETE 2>&1)

    if [[ $DELETE_OUTPUT == *"Deleted snapshot"* ]]; then
        log "Snapshot deleted ${GREEN}successfully${NC}"
    else
        log "${RED}Error during snapshot deletion: ${DELETE_OUTPUT}${NC}"
        exit 1
    fi
}

# Function to list snapshots
list_snapshots() {
    log "Listing all snapshots for database ${GREEN}${DATABASE_NAME}${NC}..."

    # Get the list of snapshots
    SNAPSHOTS=$($YB_ADMIN_PATH -master_addresses $MASTER_ADDRESSES list_snapshots 2>&1)

    if [[ $SNAPSHOTS == *"No snapshots found"* ]]; then
        log "${RED}No snapshots found.${NC}"
    else
        log "List of snapshots:"
        echo "$SNAPSHOTS"
    fi
}

# Function to display the main menu
main_menu() {
    while true; do
        # Display the main menu with numbered options
        echo -e "\nSelect an operation:"
        echo "1. Backup"
        echo "2. Restore"
        echo "3. Delete"
        echo "4. List Snapshots"
        echo "5. Exit"

        # Prompt the user for input
        read -p "Enter the number of your choice: " choice

        case $choice in
            1)
                create_snapshot
                ;;
            2)
                echo "Enter the snapshot UUID to restore:"
                read UUID
                restore_snapshot "$UUID"
                ;;
            3)
                echo "Enter the snapshot UUID to delete:"
                read UUID
                delete_snapshot "$UUID"
                ;;
            4)
                list_snapshots
                ;;
            5)
                log "Exiting script. Goodbye!"
                exit 0
                ;;
            *)
                log "${RED}Invalid option. Please select a valid number from the menu.${NC}"
                ;;
        esac

        # After the action, prompt the user to return to the main menu or exit
        echo -e "\nDo you want to return to the main menu or exit? (1 = Return, 2 = Exit)"
        read -p "Your choice: " return_choice
        if [[ "$return_choice" == "2" ]]; then
            log "Exiting script. Goodbye!"
            exit 0
        fi
    done
}

# Start the script with the main menu
main_menu