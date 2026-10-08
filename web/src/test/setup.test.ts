import { getConfig } from "@testing-library/react";

describe("test setup", () => {
  // C73
  it("async timeout is 5000 ms for every findBy query", () => {
    expect(getConfig().asyncUtilTimeout).toBe(5000);
  });
});
