# premise-pulumi

This repository contains Premise infrastructure templates. Templates compose published providers for a specific deployment. They do not implement providers or a general cloud abstraction layer.

## Templates

| Template      | Purpose                                                                           |
| ------------- | --------------------------------------------------------------------------------- |
| `gcs-bucket`  | Create a private GCS bucket and upload bound application artifacts.               |
| `neon-vercel` | Create an isolated Neon database branch and deploy a bound application to Vercel. |

Each generated environment program uses a stack per environment. Database and application resources belong to the same stack. Client-owned components can extend the program through normal language imports.

```sh
pm g cloudvoyant/premise-pulumi:gcs-bucket --answer name=environment
pm infra bind web environment --artifact dist
pm infra init --env dev
pm infra plan --env dev
pm infra apply --env dev --yes
```

Configure a named remote backend and environment values in the client manifest before initialization. The backend must already exist. Templates do not provision their own state store.

## Neon and Vercel example

The example generates `premise-tanstack-start-app` from premise-bun. It adds a small server endpoint and committed database migrations. It binds the app to the generated environment program through `pm infra bind`.

The deployment provisions database prerequisites, runs migrations and applies the full stack. The client owns migration compatibility. After a health failure, Premise invokes the app's rollback task to restore the previous deployment.

```sh
cd examples/neon-vercel
npm install --ignore-scripts
PREMISE_BIN=/path/to/pm node scaffold.mjs /path/to/client
```

The scaffold reads non-secret project and backend IDs from the environment. It does not require Neon or Vercel administration tokens. The template uses Pulumi's parameterized published Neon provider and the published Vercel provider. `pulumi install` generates the Neon SDK from pinned package metadata. Generated SDK files are not source-controlled.

## Live tests

The hosted workflow requires `PREMISE_RC_VERSION` to identify a published Premise RC. It checks out that tag's action and installs that release's binary. Configure GCP workload identity and CI-only Neon/Vercel credentials before running it.

| Test                                          | Required evidence                                                                                                              |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `mise run e2e:secrets`                        | fnox/Mise resolves a real application secret into a private env file and a child process. Administration credentials stay out. |
| `mise run e2e:gcs`                            | A generated program uploads a bound object, reuses remote state, reapplies and cleans up its bucket.                           |
| `npm run test:live` in `examples/neon-vercel` | SQL records prove migration execution. HTTP and provider identities prove app rollout and automatic rollback.                  |

The application test deploys A, migrates and deploys B, then deliberately fails a health check. Premise restores A automatically. The test queries Neon and the serving app to make sure that data survives. A separate fault exercises failed recovery.

Neon/Vercel administration operations run in CI. The GCS and application-secret tests can run locally with authorized GCP credentials. A local CLI build does not satisfy the published-RC validation gate.

## Cleanup and recovery

Tests create preview resources inside dedicated test projects. They preserve project containers and the remote state backend. They delete retained Vercel deployments by exact test-owned IDs.

Keep release records and pinned artifacts in durable storage. The tests restore records after removing their local copy. If cleanup fails, report the remaining resources before retrying. If a deployment stops unexpectedly, inspect its stack, migration history and recovery records.

## Documentation

Read Premise's [secret management guide](https://github.com/cloudvoyant/premise/blob/feature/pulumi/docs/secret-management.md) and [infrastructure guide](https://github.com/cloudvoyant/premise/blob/feature/pulumi/docs/modules/infrastructure.md). The tandem PRs must link the verified RC and hosted test runs. Replace feature-branch links with stable documentation after merge.
