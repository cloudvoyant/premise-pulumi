"""Exercise generated GCS infrastructure using a configured remote backend."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

REPOSITORY = Path(__file__).resolve().parents[1]
PM = os.environ.get("PREMISE_BIN", "pm")


def run(directory, *arguments):
    process = subprocess.run(arguments, cwd=directory, capture_output=True, text=True)
    if process.returncode:
        raise RuntimeError(f"{arguments[0]} operation failed: {process.stderr}")
    return process.stdout


def exercise():
    project = os.environ["GCP_PROJECT_ID"]
    backend = os.environ["PULUMI_BACKEND_URL"]
    if not backend.startswith("gs://"):
        raise RuntimeError("This test requires the configured remote GCS backend")
    branch = "gcs-e2e-" + os.environ.get("GITHUB_RUN_ID", "local")
    with tempfile.TemporaryDirectory(prefix="premise-gcs-") as temporary:
        client = Path(temporary) / "client"
        client.mkdir()
        run(client, "git", "init", "-q")
        run(client, PM, "init")
        run(client, PM, "g", str(REPOSITORY) + ":gcs-bucket", "--answer", "name=bucket")
        (client / "apps/sample/dist").mkdir(parents=True)
        (client / "apps/sample/dist/index.txt").write_text("premise-gcs-e2e")
        # The bucket smoke uses a small artifact fixture. The Neon/Vercel suite
        # separately generates the real TanStack application through pm g.
        manifest = f'''workspace:
  name: gcs-e2e
  schema-version: "0.2"
  providers: {{ci: github, tools: mise, tasks: mise, infra: pulumi, versioning: svu}}
  infra:
    backend: remote
    backends:
      remote: {{url: {backend}}}
    environments:
      preview:
        config:
          gcp:project: {project}
          location: {os.environ.get("GCP_BUCKET_LOCATION", "US")}
  projects:
    - name: bucket
      template: "{REPOSITORY}:gcs-bucket"
      path: infra/bucket
      infra: {{role: program, scope: environment}}
    - name: sample
      template: fixture:sample
      path: apps/sample
'''
        (client / "premise.yaml").write_text(manifest)
        run(client, PM, "infra", "bind", "sample", "bucket", "--artifact", "dist")
        run(client, "mise", "trust", str(client / "mise.toml"))
        run(client, "mise", "trust", str(client / "infra/bucket/mise.toml"))
        selection = ("--env", "preview", "--branch", branch)
        initialized = False
        try:
            run(client, PM, "infra", "init", *selection)
            initialized = True
            run(client, PM, "infra", "plan", *selection)
            run(client, PM, "infra", "apply", *selection, "--yes")
            outputs = json.loads(run(client, PM, "infra", "output", *selection))
            bucket = outputs["bucketName"]
            content = run(client, "gcloud", "storage", "cat", f"gs://{bucket}/sample/index.txt")
            assert content == "premise-gcs-e2e", "Bound artifact did not reach GCS"
            run(client, PM, "infra", "apply", *selection, "--yes")
            second = Path(temporary) / "fresh-client"
            def source_only(directory, names):
                return [name for name in names if name in {".git", ".pulumi"}
                        or (name.startswith("Pulumi.") and name.endswith(".yaml") and name != "Pulumi.yaml")]
            shutil.copytree(client, second, ignore=source_only)
            run(second, "mise", "trust", str(second / "mise.toml"))
            run(second, "mise", "trust", str(second / "infra/bucket/mise.toml"))
            run(second, PM, "infra", "init", *selection)
            assert json.loads(run(second, PM, "infra", "output", *selection))["bucketName"] == bucket
            run(second, PM, "infra", "plan", *selection)
            print("GCS preview/apply/artifact/reapply/fresh-directory state assertions passed")
        finally:
            if initialized:
                run(client, PM, "infra", "destroy", *selection)
                if "bucket" in locals():
                    result = subprocess.run(["gcloud", "storage", "buckets", "describe", f"gs://{bucket}"], capture_output=True)
                    assert result.returncode != 0, "Test bucket survived cleanup"
                print("GCS preview resources cleaned up; project and backend preserved")


if __name__ == "__main__":
    exercise()
