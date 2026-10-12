for (const button of document.querySelectorAll("[data-dialog]")) {
  button.addEventListener("click", () => {
    document.getElementById(button.dataset.dialog).showModal();
  });
}
