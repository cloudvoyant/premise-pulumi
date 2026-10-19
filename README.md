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

## Validation

`mise run test` runs Go component mocks; `mise run lint` runs Go static analysis.
The component tests check independent Neon databases and Vercel source/alias
bindings without creating cloud resources.

The live workflow checks out the exact published Premise RC and runs its
`TestInfraGCS` Go smoke test. Premise owns the small local template registry and
dummy client fixture. The test generates infrastructure, binds an artifact,
previews/applies a private GCS bucket, checks its uploaded file, reloads remote
state, and destroys the bucket. There is no JavaScript/Python E2E runner.

Configure these GitHub values for the protected `preview` environment:

| Setting                                                 | Storage   | Purpose                                                       |
| ------------------------------------------------------- | --------- | ------------------------------------------------------------- |
| `PREMISE_RC_VERSION`                                    | Variable  | Published Premise RC containing `TestInfraGCS`.               |
| `GCP_PROJECT_ID`                                        | Variable  | Dedicated test project with the Storage API enabled.          |
| `GCP_SERVICE_ACCOUNT`, `GCP_WORKLOAD_IDENTITY_PROVIDER` | Variables | Workload identity with test bucket/object and backend access. |
| `PULUMI_BACKEND_URL`                                    | Variable  | Existing GCS state bucket/prefix.                             |

Mise installs the gcloud version pinned in `mise.toml`. The live job runs `pm infra auth --ci` to exchange the runner identity for standard GCP credential files. It does not use a Google auth action. See Premise's [provider setup](https://github.com/cloudvoyant/premise/blob/v0.5.0-rc.244/docs/providers/README.md)
for the same contract on other CI systems.

The smoke test uses an empty Pulumi passphrase and relies on backend access
controls and storage encryption. No passphrase CI secret is required.

The smoke test uses only GCP OAuth/workload identity. It requires no Secret
Manager or Cloud Resource Manager permission check, Neon credentials, or Vercel
credentials. It creates a unique preview stack and always attempts cleanup,
including after a failed assertion. The state bucket and project remain.

Premise's unit tests cover migration ordering and recovery behavior with fake
subprocesses. This GCS test validates infrastructure commands against real cloud
resources; it does not claim live database migration or application rollback
coverage.

## Documentation

Read Premise's [secret management guide](https://github.com/cloudvoyant/premise/blob/feature/pulumi/docs/secret-management.md) and [infrastructure guide](https://github.com/cloudvoyant/premise/blob/feature/pulumi/docs/modules/infrastructure.md). The tandem PRs must link the verified RC and hosted test runs. Replace feature-branch links with stable documentation after merge.

## Provider guides

GCP, Neon, and Vercel have separate setup guides and separate template contracts:

- [GCP](docs/providers/gcp.md) configures the live bucket test.
- [Neon](docs/providers/neon.md) configures `neon-db` administration access.
- [Vercel](docs/providers/vercel.md) configures `vercel-app` administration access.

Each guide uses the same Tools, Local authentication, CI authentication, Variables, Permissions, and References sections. Combining their components in a client program does not combine their credentials or providers. Fnox remains local. CI receives application values from CI secrets.
