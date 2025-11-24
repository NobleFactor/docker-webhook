#!/usr/bin/env gawk -f

BEGIN {
    # Uses gawk -i inplace for in-place editing
    
    # Array of keys to potentially update
    keys["AZURE_CLIENT_ID"] = azure_client_id
    keys["AZURE_TENANT_ID"] = azure_tenant_id
    keys["WEBHOOK_KEYVAULT_URL"] = keyvault_url
    keys["WEBHOOK_LOCATION"] = location
    keys["WEBHOOK_SECRET_NAME"] = secret_name
    keys["WEBHOOK_SP"] = grant_access_to
    keys["WEBHOOK_TOKEN_REFRESH_WINDOW"] = "5m"
    keys["WEBHOOK_TOKEN_TTL"] = "24h"
    
    # Only include Azure and SP keys if azure_client_id is set
    if (azure_client_id == "") {
        delete keys["AZURE_CLIENT_ID"]
        delete keys["AZURE_TENANT_ID"]
        delete keys["WEBHOOK_SP"]
    }
}

{
    # Process each line of the input file
    updated = 0
    for (key in keys) {
        if ($0 ~ "^export " key "=") {
            print "export " key "=" keys[key]
            delete keys[key]  # Mark as updated to avoid appending later
            updated = 1
            break
        }
    }
    if (!updated) {
        print
    }
}

END {
    # Build a sorted array of key names
    n = asorti(keys, sorted_keys)

    # Append any keys that weren't found in the file
    for (i = 1; i <= n; i++) {
        key = sorted_keys[i]
        print "export " key "=" keys[key]
    }
}
