export const goEnv = {
  GOPROXY:
    "https://pkg.vpn.patinanetwork.org/go/go,https://proxy.golang.org,direct",
  GONOSUMDB: "patinanetwork.org",
  ...process.env,
};
