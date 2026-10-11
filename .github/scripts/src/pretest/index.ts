import { $ } from "bun";

import { goEnv } from "../toolchain";

async function main() {
  const unformatted = (await $`gofmt -l cmd internal frontend/embed.go`.text()).trim();
  if (unformatted) {
    throw new Error(`gofmt found unformatted files:\n${unformatted}`);
  }
  await $`go vet ./...`.env(goEnv);
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((e) => {
    console.error(e);
    process.exit(1);
  });
