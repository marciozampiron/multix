#!/usr/bin/env bash
# File: scripts/e2e/oci_smoke.sh
# Company: Hassan
# Creator: Zamp
# Created: 16/03/2026
# Updated: 16/03/2026
# Purpose: End-to-end smoke validation for OCI auth flows in MULTIX.

set -euo pipefail

# --- Configuration ---
MULTIX_BIN="${MULTIX_BIN:-build/multix}"
OCI_CONFIG_FILE="${OCI_CONFIG_FILE:-$HOME/.oci/config}"

# --- Helper Functions ---
log_info() {
    printf '[\033[0;34mINFO\033[0m] %s\n' "$1"
}

log_pass() {
    printf '[\033[0;32mPASS\033[0m] %s\n' "$1"
}

log_fail() {
    printf '[\033[0;31mFAIL\033[0m] %s\n' "$1"
    exit 1
}

check_no_panic() {
    local output="$1"
    local contextual_msg="$2"
    if echo "$output" | grep -E "^panic:|^goroutine "; then
        log_fail "Panic detected during $contextual_msg. Output:\n$output"
    fi
}

# run_command_capture runs a command, captures combined stdout/stderr,
# and captures the exit code.
# The exit code is stored in the global variable CMD_EXIT_CODE.
# The output is stored in the global variable CMD_OUTPUT.
run_command_capture() {
    # Using a temporary file to capture output reliably across bash versions
    # without subshell masking of exit codes.
    local tmp_out
    tmp_out=$(mktemp)
    
    # Run the command, redirecting stderr to stdout, and then to the temp file
    # We turn off exit-on-error temporarily just for this execution
    set +e
    "$@" > "$tmp_out" 2>&1
    CMD_EXIT_CODE=$?
    set -e
    
    CMD_OUTPUT=$(cat "$tmp_out")
    rm -f "$tmp_out"
}

# --- Execution ---

log_info "Building MULTIX binary to $MULTIX_BIN..."
go build -o "$MULTIX_BIN" ./cmd/multix

has_credentials=false
if [ -f "$OCI_CONFIG_FILE" ]; then
    log_info "OCI config detected: yes ($OCI_CONFIG_FILE)"
    has_credentials=true
else
    log_info "OCI config detected: no"
fi

# ==========================================
# 1. Test: auth validate
# ==========================================
log_info "Running: $MULTIX_BIN auth validate --provider oci"
run_command_capture "$MULTIX_BIN" auth validate --provider oci

check_no_panic "$CMD_OUTPUT" "auth validate"

if [ -z "$CMD_OUTPUT" ]; then
    log_fail "auth validate returned empty output."
fi

if [ "$has_credentials" = true ]; then
    if [ "$CMD_EXIT_CODE" -eq 0 ]; then
        log_pass "auth validate succeeded with exit code 0."
    elif echo "$CMD_OUTPUT" | grep -qiE "NotAuthorizedOrNotFound|not authorized|401|403"; then
        log_pass "auth validate hit Identity API and was rejected properly (status $CMD_EXIT_CODE)."
    elif echo "$CMD_OUTPUT" | grep -qiE "bad configuration|missing tenancy|can not create client|failed to retrieve"; then
        log_fail "auth validate failed due to bad local config even though config file exists. Output:\n$CMD_OUTPUT"
    else
        log_fail "auth validate failed unexpectedly (status $CMD_EXIT_CODE). Output:\n$CMD_OUTPUT"
    fi
else
    if [ "$CMD_EXIT_CODE" -ne 0 ]; then
        if echo "$CMD_OUTPUT" | grep -qiE "bad configuration|not configured|missing tenancy|failed to retrieve|can not create client"; then
            log_pass "auth validate rejected missing/bad configuration gracefully (status $CMD_EXIT_CODE)."
        else
            log_fail "auth validate failed with unexpected output (status $CMD_EXIT_CODE). Output:\n$CMD_OUTPUT"
        fi
    else
        log_fail "auth validate returned exit code 0 despite missing credentials. Output:\n$CMD_OUTPUT"
    fi
fi

# ==========================================
# 2. Test: auth whoami
# ==========================================
log_info "Running: $MULTIX_BIN auth whoami --provider oci"
run_command_capture "$MULTIX_BIN" auth whoami --provider oci

check_no_panic "$CMD_OUTPUT" "auth whoami"

if [ -z "$CMD_OUTPUT" ]; then
    log_fail "auth whoami returned empty output."
fi

if [ "$has_credentials" = true ]; then
    if [ "$CMD_EXIT_CODE" -eq 0 ]; then
        log_pass "auth whoami succeeded with exit code 0."
    elif echo "$CMD_OUTPUT" | grep -qiE "NotAuthorizedOrNotFound|not authorized|401|403"; then
        log_pass "auth whoami hit Identity API and was rejected properly (status $CMD_EXIT_CODE)."
    elif echo "$CMD_OUTPUT" | grep -qiE "bad configuration|missing tenancy|can not create client|failed to retrieve"; then
        log_fail "auth whoami failed due to bad local config even though config file exists. Output:\n$CMD_OUTPUT"
    else
        log_fail "auth whoami failed unexpectedly (status $CMD_EXIT_CODE). Output:\n$CMD_OUTPUT"
    fi
else
    if [ "$CMD_EXIT_CODE" -ne 0 ]; then
        if echo "$CMD_OUTPUT" | grep -qiE "bad configuration|not configured|missing tenancy|failed to retrieve|can not create client|missing user"; then
            log_pass "auth whoami rejected missing/bad configuration gracefully (status $CMD_EXIT_CODE)."
        else
            log_fail "auth whoami failed with unexpected output (status $CMD_EXIT_CODE). Output:\n$CMD_OUTPUT"
        fi
    else
        log_fail "auth whoami returned exit code 0 despite missing credentials. Output:\n$CMD_OUTPUT"
    fi
fi

log_pass "OCI smoke test completed safely."
exit 0
