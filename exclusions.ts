/**
 * See https://github.com/Patina-Network/hello-world-grpc-client-go/blob/main/.github/scripts/src/test/index.ts
 * for test exclusion usages.
 */

const baseDir = ".";

export const exclusions = [
  `${baseDir}/config/**`,
  "cmd/server/main.go",
  "internal/config/server.go",
  "frontend/embed.go",
  "internal/httpserve/health.go",
  "internal/httpserve/metrics.go",
];
