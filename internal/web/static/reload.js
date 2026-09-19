const url = new URL("/api/reload", origin);

if (globalThis.INSTANCE_ID) {
  url.searchParams.set("instance_id", globalThis.INSTANCE_ID);
}

const source = new EventSource(url);
source.addEventListener("reload", (event) => {
  location.reload();
});
