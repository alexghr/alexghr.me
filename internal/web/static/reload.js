const source = new EventSource("/api/reload");
source.addEventListener("reload", (event) => {
  console.log(event);
});
