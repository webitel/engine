#!/bin/bash
#
# Engine-specific postinst setup.
#
# Run by the generic Debian postinst (webitel/reusable-configs) BEFORE any
# unit is started. Sourced under `set -e`, so a failure aborts the install.

I18N_DIR=/usr/share/webitel/engine/i18n
[ -d "$I18N_DIR" ] || install -d -o "$USER_NAME" -g "$GROUP_NAME" -m 0755 "$I18N_DIR"

# Engine reads the certificates FreeSWITCH shares through its group.
if getent group freeswitch >/dev/null 2>&1; then
    adduser "$USER_NAME" freeswitch
fi
