# premise-pulumi

This repository contains Premise infrastructure templates. Templates compose published providers for a specific deployment. They do not implement providers or a general cloud abstraction layer.

## Templates

| Template      | Purpose                                                              |
| ------------- | -------------------------------------------------------------------- |
| `gcs-bucket`  | Create a private GCS bucket and upload bound application artifacts.  |
| `environment` | Create a client-owned Go environment program.                        |
| `neon-db`     | Scaffold a Go component for a branch, endpoint, role, and database.  |
| `vercel-app`  | Scaffold a Go component for a bound app deployment and stable alias. |

Each generated environment program uses a stack per environment. Database and application resources belong to the same stack. Client-owned components can extend the program through normal language imports.

```sh
pm g cloudvoyant/premise-pulumi:gcs-bucket --answer name=environment
pm infra bind web environment --artifact dist
pm infra init --env dev
pm infra plan --env dev
pm infra apply --env dev --yes
```

Configure a named remote backend and environment values in the client manifest before initialization. The backend must already exist. Templates do not provision their own state store.

## Component composition

Generate an environment program and project components with `pm g`:

```sh
pm g cloudvoyant/premise-pulumi:environment --answer name=environment
pm g cloudvoyant/premise-pulumi:neon-db --answer name=database
pm g cloudvoyant/premise-pulumi:vercel-app --answer name=web-deployment
pm g cloudvoyant/premise-pulumi:vercel-app --answer name=admin-deployment
```

Import the rendered Go modules from the client-owned environment program. Each
component constructor accepts its own resource name and project inputs. The same
Vercel template can be rendered twice or instantiated twice. Bind each app with
`pm infra bind`, then use `pm infra init/plan/apply`. Pulumi owns provider settings;
`PULUMI_BACKEND_URL` can come from Mise/fnox. Named manifest backends are optional.

The embedded client fixture in Premise shows the complete Go composition and
optional migration/recovery tasks. Its live test uses the parameterized published
Neon provider and the published Vercel provider. Mise generates the Go Neon SDK
from pinned package metadata; generated SDK files are not source-controlled.
There is no product setup script or `examples` directory in this registry.

## Live tests

This branch requires Premise `v0.5.0-rc.227` or a later release with the same infrastructure, deployment and secret interfaces. The matching Premise change is [PR #20](https://github.com/cloudvoyant/premise/pull/20); merge it before these templates.

The hosted workflow requires `PREMISE_RC_VERSION` to identify a published Premise RC. It checks out that tag's action and installs that release's binary. The harness executes the action's published deployment-step scripts, including remote record restore and persistence. Configure GCP workload identity and CI-only Neon/Vercel credentials before running it.

| Test                                       | Required evidence                                                                                                              |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `mise run e2e:secrets`                     | fnox/Mise resolves a real application secret into a private env file and a child process. Administration credentials stay out. |
| `mise run e2e:gcs`                         | A generated program uploads a bound object, reuses remote state, reapplies and cleans up its bucket.                           |
| `TestPulumiLive` in the Premise repository | SQL records prove migration execution. HTTP and provider identities prove app rollout and automatic rollback.                  |

Configure these GitHub repository values before a feature-branch push:

| Setting                                                                   | GitHub storage | Purpose                                                                        |
| ------------------------------------------------------------------------- | -------------- | ------------------------------------------------------------------------------ |
| `PREMISE_RC_VERSION`                                                      | Variable       | Exact verified RC tag; currently `v0.5.0-rc.227`.                              |
| `GCP_PROJECT_ID`, `GCP_SERVICE_ACCOUNT`, `GCP_WORKLOAD_IDENTITY_PROVIDER` | Variables      | Workload identity for the dedicated GCP project.                               |
| `PULUMI_BACKEND_URL`                                                      | Variable       | Existing GCS state backend.                                                    |
| `NEON_PROJECT_ID`, `NEON_PARENT_BRANCH_ID`, `NEON_REGION_ID`              | Variables      | Existing dedicated Neon project and parent branch.                             |
| `VERCEL_PROJECT_ID`, `VERCEL_TEAM_ID`                                     | Variables      | Existing dedicated Vercel project and team.                                    |
| `NEON_API_KEY`, `VERCEL_TOKEN`                                            | CI secrets     | Provider administration; supplied only to the live deployment step.            |
| `PULUMI_CONFIG_PASSPHRASE`                                                | CI secret      | State encryption; retain a private recovery copy.                              |
| `PREMISE_E2E_BYPASS_SECRET`                                               | CI secret      | Automation access when the dedicated Vercel project has deployment protection. |

Enable the GCP storage and Secret Manager APIs. Give workload identity access to the state bucket, test bucket creation and the dedicated `premise-e2e-app-marker` secret, plus Secret Manager creation/deletion for branch-owned fixture secrets. The marker is an application value in Secret Manager, not a copied CI secret. Configure the GitHub `preview` environment to permit the test feature branches. Keep Vercel deployment protection enabled and create a project-scoped automation bypass for HTTP tests.

The embedded client fixture in Premise generates these Go components with `pm g`, renders the Vercel template twice for two apps, and composes them with ordinary Go imports. There is no generated JavaScript deploy script. The application test deploys A, migrates and deploys B, then deliberately fails a health check. Premise restores A automatically. The test queries Neon and the serving app to make sure that data survives. A separate fault exercises failed recovery.

Neon/Vercel administration operations run in CI. The GCS and application-secret tests can run locally with authorized GCP credentials. A local CLI build does not satisfy the published-RC validation gate.

## Cleanup and recovery

Tests create preview resources inside dedicated test projects. They preserve project containers and the remote state backend. They delete retained Vercel deployments by exact test-owned IDs. After resource cleanup, they archive non-secret recovery records by CI run and clear active preview records. A later test must not restore an app from a destroyed preview.

Keep release records and pinned artifacts in durable storage. The tests restore records after removing their local copy. If cleanup fails, report the remaining resources before retrying. If a deployment stops unexpectedly, inspect its stack, migration history and recovery records.

## Documentation

Read Premise's [secret management guide](https://github.com/cloudvoyant/premise/blob/feature/pulumi/docs/secret-management.md) and [infrastructure guide](https://github.com/cloudvoyant/premise/blob/feature/pulumi/docs/modules/infrastructure.md). The tandem PRs must link the verified RC and hosted test runs. Replace feature-branch links with stable documentation after merge.
