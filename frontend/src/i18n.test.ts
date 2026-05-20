import { describe, expect, it } from "vitest";

import { t } from "./i18n";

describe("i18n", () => {
  it("uses Chinese as the default-facing content set", () => {
    expect(t("zh-CN", "title")).toBe("里程碑驱动项目管理");
  });

  it("supports English fallback content", () => {
    expect(t("en-US", "backend")).toBe("Backend: Golang");
  });
});
