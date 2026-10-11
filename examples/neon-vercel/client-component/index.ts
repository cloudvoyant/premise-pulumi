import * as pulumi from "@pulumi/pulumi";

/** A client-owned component imported through ordinary TypeScript modules. */
export class ProjectIdentity extends pulumi.ComponentResource {
  readonly label: string;
  constructor(stack: string) {
    super("client:identity:Project", "client-identity");
    this.label = stack;
    this.registerOutputs({ label: stack });
  }
}
