#!/bin/bash

# Function to display help
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo "Commands:"
    echo "  migrate         - Apply database migrations"
    echo "  make-admin      - Make a user admin (requires email)"
    echo "  help            - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 migrate"
    echo "  $0 make-admin user@example.com"
    echo ""
}

# Function to apply migrations
apply_migrations() {
    echo "Applying database migration..."
    export PGPASSWORD='S#nhaP1xelcraft'
    psql -U pixelcraft_user -d pixelcraft -h localhost -f backend/database/migration_subscriptions.sql
    echo "Migration applied."
}

# Function to make user admin
make_admin() {
    local email="$1"

    if [ -z "$email" ]; then
        echo "Error: Email is required"
        echo "Usage: $0 make-admin <email>"
        exit 1
    fi

    echo "Setting user with email '$email' as admin..."
    export PGPASSWORD='S#nhaP1xelcraft'

    # Check if user exists
    local user_exists
    user_exists=$(psql -U pixelcraft_user -d pixelcraft -h localhost -t -c "SELECT COUNT(*) FROM users WHERE email = '$email';" 2>/dev/null)

    if [ "$user_exists" = "0" ]; then
        echo "Error: User with email '$email' not found"
        exit 1
    fi

    # Update user to be admin
    psql -U pixelcraft_user -d pixelcraft -h localhost -c "UPDATE users SET is_admin = true WHERE email = '$email';"

    if [ $? -eq 0 ]; then
        echo "User with email '$email' is now an admin."
    else
        echo "Error: Failed to update user admin status"
        exit 1
    fi
}

# Main script logic
case "$1" in
    migrate)
        apply_migrations
        ;;
    make-admin)
        make_admin "$2"
        ;;
    help|"")
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        echo "Use '$0 help' for usage information."
        exit 1
        ;;
esac
