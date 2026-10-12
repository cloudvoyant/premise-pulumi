"""Check the live fixture's GCP permissions before it creates provider resources."""
import json
import os
import subprocess
import urllib.request


def check():
    token = subprocess.run(
        ["gcloud", "auth", "print-access-token"], capture_output=True, text=True, check=True
    ).stdout.strip()
    permissions = [
        "secretmanager.secrets.create",
        "secretmanager.secrets.delete",
        "secretmanager.versions.add",
    ]
    request = urllib.request.Request(
        f"https://cloudresourcemanager.googleapis.com/v1/projects/{os.environ['GCP_PROJECT_ID']}:testIamPermissions",
        data=json.dumps({"permissions": permissions}).encode(),
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
    )
    with urllib.request.urlopen(request) as response:
        granted = json.load(response).get("permissions", [])
    missing = set(permissions) - set(granted)
    if missing:
        raise SystemExit("Live fixture needs dedicated-project permissions: " + ", ".join(sorted(missing)) + "; no infrastructure was created")
    print("Live fixture Secret Manager permissions verified")


if __name__ == "__main__":
    check()
