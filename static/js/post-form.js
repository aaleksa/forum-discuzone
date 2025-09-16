document.addEventListener("DOMContentLoaded", () => {
  const textarea = document.getElementById("content");
  const counter = document.getElementById("char-counter");
  const maxLength = parseInt(textarea.getAttribute("maxlength"), 10);

  if (textarea && counter) {
    const updateCounter = () => {
      const currentLength = textarea.value.length;
      counter.textContent = `${currentLength}/${maxLength} symbols`;
    };

    textarea.addEventListener("input", updateCounter);
    updateCounter(); // Show the current value when loading
  }
});
