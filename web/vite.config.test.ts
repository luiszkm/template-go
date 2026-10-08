import config from "./vite.config";

describe("vite dev server", () => {
  it("proxies backend paths to the api server", () => {
    const proxy = config.server?.proxy ?? {};
    for (const path of ["/api", "/healthz", "/readyz"]) {
      expect(proxy[path]).toBe("http://localhost:8080");
    }
  });
});
