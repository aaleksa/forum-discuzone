function togglePassword() {
  const passwordInput = document.getElementById("password");
  const icon = document.getElementById("eye-icon");

  const isPassword = passwordInput.type === "password";
  passwordInput.type = isPassword ? "text" : "password";

  // Change the entire SVG icon
  if (isPassword) {
    icon.setAttribute("viewBox", "0 0 24 24");
    icon.innerHTML = `
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
        d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
    `;
  } else {
    icon.setAttribute("viewBox", "0 0 24 24");
    icon.innerHTML = `
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
        d="M15 12a3 3 0 11-6 0 3 3 0 016 0zm7.242 0a10.97 10.97 0 01-1.383 2.12M12 5c3.866 0 7.09 2.91 8.485 7-1.395 4.09-4.619 7-8.485 7s-7.09-2.91-8.485-7C4.91 7.91 8.134 5 12 5z" />
    `;
  }
}