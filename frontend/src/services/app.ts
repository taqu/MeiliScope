import * as AppServiceBindings from "../../bindings/github.com/taqu/meiliscope/internal/services/appservice.js";
import type { ApplicationInfo, RuntimeInfo } from "../../bindings/github.com/taqu/meiliscope/internal/services/models.js";

export type { ApplicationInfo, RuntimeInfo };

export const AppService = {
  getApplicationInfo: (): Promise<ApplicationInfo> =>
    AppServiceBindings.GetApplicationInfo(),
  getApplicationVersion: (): Promise<string> =>
    AppServiceBindings.GetApplicationVersion(),
  getRuntimeInformation: (): Promise<RuntimeInfo> =>
    AppServiceBindings.GetRuntimeInformation(),
  getFormattedUptime: (): Promise<string> =>
    AppServiceBindings.GetFormattedUptime(),
};
