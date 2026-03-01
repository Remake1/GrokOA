import "@fontsource-variable/open-sans";
import "@fontsource-variable/lora";
import "@fontsource-variable/jetbrains-mono";
import { createApp } from "vue";
import "./style.css";
import App from "./App.vue";
import router from "./router";

async function bootstrap() {
    if (import.meta.env.DEV) {
        const { worker } = await import("./mocks/browser");
        await worker.start({ onUnhandledRequest: "bypass" });
    }

    createApp(App).use(router).mount("#app");
}

bootstrap();
