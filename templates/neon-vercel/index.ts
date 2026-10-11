import * as pulumi from "@pulumi/pulumi";
import * as path from "node:path";
import { createDatabase } from "./database";
import { createApplication } from "./application";

const config = new pulumi.Config();
const context = new pulumi.Config("premise");
const stack = context.require("stack");
const workspaceRoot = path.resolve("../..");
const artifacts = config.requireObject<Record<string, string>>("artifacts");
const app = config.get("app") ?? "web";
const source = artifacts[app];
if (!source) throw new Error(`App ${app} must be bound with pm infra bind`);
const sourceDirectory = path.resolve(source);
if (!sourceDirectory.startsWith(workspaceRoot + path.sep)) {
  throw new Error("Bound application source must stay inside the workspace");
}
const database = createDatabase(
  config.require("neonProjectId"),
  config.require("neonParentBranchId"),
  stack,
);
const alias = `premise-${stack.slice(-10)}-${config.require("vercelTeamSlug")}.vercel.app`;
const appSecrets = JSON.parse(process.env.PM_APP_SECRETS ?? "{}")[app] ?? {};
const application = createApplication(
  config.require("vercelProjectId"),
  config.require("vercelTeamId"),
  alias,
  sourceDirectory,
  database.url,
  process.env.PM_RELEASE ?? config.get("release") ?? "unversioned",
  appSecrets,
);

export const databaseURL = database.url;
export const databaseBranchId = database.branch.id;
export const deploymentId = application.deployment.id;
export const appURL = pulumi.interpolate`https://${application.route.alias}`;
export const aliasName = application.route.alias;
