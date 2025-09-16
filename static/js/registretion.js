document.addEventListener("DOMContentLoaded", function () {
  const form = document.getElementById("registerForm");
  const successModal = document.getElementById("successModal");
  const errorModal = document.getElementById("errorModal");

  form.addEventListener("submit", async function (e) {
    e.preventDefault();

    const formData = new FormData(form);
    const password = formData.get("password");
    const confirmPassword = formData.get("confirm_password");

    if (password !== confirmPassword) {
      showModal("error", "Passwords do not match.");
      return;
    }

    if (password.length < 8) {
      showModal("error", "Password must contain at least 8 characters");
      return;
    }

    try {
      const response = await fetch("/register-submit", {
        method: "POST",
        body: formData,
        headers: { Accept: "application/json" },
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || "Registration error");
      }

      showModal("success", result.message || "Registration successful!");
      form.reset();
    } catch (err) {
      showModal("error", err.message || "An error occurred.");
    }
  });

  function showModal(type, message) {
    if (type === "success") {
      successModal.querySelector("p").textContent = message;
      successModal.style.display = "flex";
      setTimeout(() => (window.location.href = "/login"), 3000);
    } else {
      errorModal.querySelector("p").textContent = message;
      errorModal.style.display = "flex";
    }
  }

  window.closeModal = () => (successModal.style.display = "none");
  window.closeErrorModal = () => (errorModal.style.display = "none");

  window.addEventListener("click", function (event) {
    if (event.target === successModal) closeModal();
    if (event.target === errorModal) closeErrorModal();
  });
});