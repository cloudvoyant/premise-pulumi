import * as pulumi from "@pulumi/pulumi";
import * as neon from "@pulumi/neon";

/** A branch, role, database and small compute endpoint in an existing Neon project. */
export function createDatabase(
  projectId: string,
  parentId: string,
  stack: string,
) {
  const branch = new neon.Branch("db-branch", {
    projectId,
    parentId,
    name: `premise-${stack}`,
  });
  const role = new neon.Role("db-role", {
    projectId,
    branchId: branch.id,
    name: "premise_app",
  });
  const database = new neon.Database("db-database", {
    projectId,
    branchId: branch.id,
    name: "premise_app",
    ownerName: role.name,
  });
  const endpoint = new neon.Endpoint("db-endpoint", {
    projectId,
    branchId: branch.id,
    type: "read_write",
    autoscalingLimitMinCu: 0.25,
    autoscalingLimitMaxCu: 0.25,
    suspendTimeoutSeconds: 60,
  });
  const url = pulumi.secret(
    pulumi
      .all([role.name, role.password, endpoint.host, database.name])
      .apply(
        ([user, password, host, name]) =>
          `postgresql://${encodeURIComponent(user)}:${encodeURIComponent(password)}@${host}/${name}?sslmode=require`,
      ),
  );
  return { branch, role, database, endpoint, url };
}
