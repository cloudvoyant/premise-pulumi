import * as pulumi from "@pulumi/pulumi";
import * as vercel from "@pulumiverse/vercel";
import * as path from "node:path";

/** Deploy bound immutable source into an existing Vercel project. */
export function createApplication(
  projectId: string,
  teamId: string,
  alias: string,
  sourceDirectory: string,
  databaseURL: pulumi.Output<string>,
  release: string,
  appSecrets: Record<string, string>,
) {
  const directory = vercel.getProjectDirectoryOutput({ path: sourceDirectory });
  const files = directory.files.apply((entries) => {
    for (const file of Object.keys(entries)) {
      const parts = file.replaceAll("\\", "/").split("/");
      if (
        parts.some(
          (part) =>
            part.startsWith(".env") ||
            [
              ".pulumi",
              ".pulumi-state",
              ".premise",
              "node_modules",
              "test-results",
              ".git",
              ".npmrc",
              "fnox.toml",
            ].includes(part),
        )
      )
        throw new Error(
          "Vercel source includes a forbidden secret/state/dependency file; fix .vercelignore",
        );
    }
    return entries;
  });
  const deployment = new vercel.Deployment(
    "app-deployment",
    {
      projectId,
      teamId,
      files,
      pathPrefix: path.resolve(sourceDirectory),
      production: false,
      deleteOnDestroy: true,
      environment: pulumi.secret({
        ...appSecrets,
        DATABASE_URL: databaseURL,
        PREMISE_RELEASE: release,
      }),
      meta: { premiseRelease: release, premiseStack: pulumi.getStack() },
      projectSettings: {
        framework: "tanstack-start",
        buildCommand: "bun run build",
        installCommand: "bun install",
      },
    },
    { retainOnDelete: true },
  );
  const route = new vercel.Alias("app-alias", {
    teamId,
    alias,
    deploymentId: deployment.id,
  });
  return { deployment, route };
}
