import { SonarScannerClient } from "@tahminator/pipeline";
import { $ } from "bun";

import { exclusions } from "../../../../exclusions";
import { goEnv } from "../toolchain";

const sourceDirs = "cmd,internal,frontend";
const testFiles = "**/*_test.go";

async function main() {
  const { sonarToken } = parseCiEnv(process.env);

  const sonarClient = new SonarScannerClient({
    auth: {
      token: sonarToken,
    },
    scan: {
      additionalArgs: {
        "go.coverage.reportPaths": "coverage.out",
        tests: sourceDirs,
        "test.inclusions": testFiles,
        exclusions: testFiles,
        "coverage.exclusions": `${exclusions}`,
      },
      organization: "patina-network",
      sourceCodeDir: sourceDirs,
      projectKey: "Patina-Network_hello-world-grpc-client-go",
    },
    run: {
      runTestsCmd:
        $`go test -race -count=1 -coverprofile=coverage.out ./...`.env(goEnv),
    },
  });

  await sonarClient.runTests();
  await sonarClient.uploadTestCoverage();
}

function parseCiEnv(ciEnv: Record<string, string | undefined>) {
  const sonarToken = (() => {
    const v = ciEnv["SONAR_TOKEN"];
    if (!v) {
      throw new Error("Missing SONAR_TOKEN from .env.ci");
    }
    return v;
  })();

  return { sonarToken };
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((e) => {
    console.error(e);
    process.exit(1);
  });
