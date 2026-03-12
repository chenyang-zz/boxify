import { emptyPluginConfigSchema } from "openclaw/plugin-sdk";

import { createBoxifyNativeChannel } from "./channel.js";
import { PLUGIN_ID } from "./constants.js";

const runtimeRef = { current: null };

const plugin = {
  id: PLUGIN_ID,
  name: "Boxify",
  description: "将 Boxify 作为 OpenClaw 原生本地聊天通道接入",
  configSchema: emptyPluginConfigSchema(),
  register(api) {
    runtimeRef.current = api.runtime;
    api.registerChannel({ plugin: createBoxifyNativeChannel(runtimeRef) });
  },
};

export default plugin;
