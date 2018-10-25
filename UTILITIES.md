# Utility code / scripts

### Deploy a release

    ansible-playbook -e "build_sha=581b5f18e272048624ea5abfe133abd815c0354c" deploy.yml -i hosts

### Bucket lifecycle rule config

```json
{
    "Rules": [
        {
            "Expiration": {
                "Days": 30
            },
            "Filter": {
                "Prefix": "rig/SNAPSHOT/"
            },
            "ID": "Snapshots",
            "Status": "Enabled"
        },
        {
            "Expiration": {
                "Days": 3
            },
            "Filter": {
                "Prefix": "rig/LOG/"
            },
            "ID": "Logs",
            "Status": "Enabled"
        }
    ]
}
```