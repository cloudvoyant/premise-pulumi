"""Prove fnox app secret delivery without exposing administration credentials."""
import os
from pathlib import Path
import subprocess
import tempfile

PM = os.environ.get("PREMISE_BIN", "pm")


def run(root, *args):
    result = subprocess.run(args, cwd=root, capture_output=True, text=True)
    if result.returncode:
        raise RuntimeError(f"{args[0]} secret test command failed")
    return result.stdout


def exercise():
    with tempfile.TemporaryDirectory(prefix="premise-secrets-") as directory:
        root = Path(directory)
        app = root / "apps/web"
        app.mkdir(parents=True)
        run(root, "git", "init", "-q")
        (root / ".gitignore").write_text(".env.local\n")
        (root / "premise.yaml").write_text('''workspace:
  name: application-secret-test
  schema-version: "0.2"
  providers: {ci: github, tools: mise, tasks: mise, infra: pulumi, versioning: svu}
  projects:
    - name: web
      template: fixture:web
      path: apps/web
      secrets:
        config: fnox.toml
        profiles: {dev: dev}
        keys: [APP_MARKER]
''')
        (app / "fnox.toml").write_text(f'''[providers.gcp]
type = "gcp-sm"
project = "{os.environ['GCP_PROJECT_ID']}"
prefix = "premise-e2e-"
[profiles.dev.secrets]
APP_MARKER = {{ provider = "gcp", value = "app-marker" }}
''')
        run(root, PM, "secrets", "pull", "--env", "dev", "--project", "web")
        destination = app / ".env.local"
        assert destination.stat().st_mode & 0o777 == 0o600
        original = destination.read_bytes()
        assert b"APP_MARKER=" in original
        assert b"VERCEL" not in original and b"NEON_API_KEY" not in original
        run(root, "git", "check-ignore", "--quiet", str(destination))
        result = subprocess.run([PM, "secrets", "pull", "--env", "prod", "--project", "web"], cwd=root, capture_output=True)
        assert result.returncode != 0
        assert destination.read_bytes() == original, "Failed pull replaced the existing file"
        run(root, PM, "secrets", "exec", "--env", "dev", "--project", "web", "--", "python3", "-c", "import os; assert os.environ['APP_MARKER']; assert not os.getenv('VERCEL_TOKEN'); assert not os.getenv('NEON_API_KEY'); assert not os.getenv('PULUMI_CONFIG_PASSPHRASE')")
        print("Live fnox/Mise local-file and process delivery, permissions, no fallback and administration separation passed")


if __name__ == "__main__":
    exercise()
