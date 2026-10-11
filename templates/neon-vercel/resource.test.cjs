require("ts-node/register");
const assert = require("node:assert/strict");
const pulumi = require("@pulumi/pulumi");
const resources = [];
pulumi.runtime.setMocks(
  {
    newResource: (args) => {
      resources.push(args);
      return {
        id: `${args.name}-id`,
        state: {
          ...args.inputs,
          password: "mock-password-only",
          host: "mock.neon.invalid",
        },
      };
    },
    call: (args) =>
      args.token.includes("getProjectDirectory")
        ? {
            files: { "package.json": "test-file-metadata" },
            path: args.inputs.path,
          }
        : args.inputs,
  },
  "example",
  "preview-test",
  false,
);

pulumi.runtime
  .runInPulumiStack(async () => {
    const { createDatabase } = require("./database");
    const { createApplication } = require("./application");
    const db = createDatabase("test-project", "parent-branch", "preview-test");
    const app = createApplication(
      "test-vercel",
      "test-team",
      "test-alias.vercel.app",
      "/test/source",
      db.url,
      "release-A",
      {},
    );
    await Promise.all([
      db.url.promise(),
      app.deployment.id.promise(),
      app.route.alias.promise(),
    ]);
    assert.equal(await pulumi.isSecret(db.url), true);
    for (const name of [
      "db-branch",
      "db-role",
      "db-database",
      "db-endpoint",
      "app-deployment",
      "app-alias",
    ]) {
      assert.ok(
        resources.some((resource) => resource.name === name),
        `resource ${name} was not composed`,
      );
    }
    const deployment = resources.find(
      (resource) => resource.name === "app-deployment",
    );
    assert.equal(deployment.inputs.projectId, "test-vercel");
    assert.equal(deployment.inputs.meta.premiseRelease, "release-A");
    assert.equal(
      resources.find((resource) => resource.name === "db-branch").inputs
        .parentId,
      "parent-branch",
    );
  })
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
  });
