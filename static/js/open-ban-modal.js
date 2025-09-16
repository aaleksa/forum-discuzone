document.addEventListener("DOMContentLoaded", function() {
  const form = document.getElementById("loginForm");
  const errorModal = document.getElementById("errorModal");
  const errorText = document.getElementById("errorText");
  const submitBtn = form.querySelector('button[type="submit"]');

  form.addEventListener("submit", async function(e) {
    e.preventDefault();
    
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = "Processing...";

    try {
      const formData = new FormData(form);
      const data = {
        email: formData.get("email"),
        password: formData.get("password")
      };

      const response = await fetch("/login-submit", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Accept": "application/json"
        },
        body: JSON.stringify(data)
      });

      const result = await response.json();

      if (!response.ok) {
        // Error handling (ban, invalid data, etc.)
        if (response.status === 403 && result.error === "Account banned") {
          showBanModal(result);
        } else {
          throw new Error(result.error || "Login error");
        }
        return;
      }

      // Успішний вхід
      window.location.href = result.redirect || "/user_page";
      
    } catch (error) {
      showErrorModal(error.message);
      console.error("Login error:", error);
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = originalText;
    }
  });

 // Function to display a modal with ban details
  function showBanModal(banData) {
    let banMessage = `
      <strong>${banData.message}</strong><br><br>
      <div style="text-align: left;">
        <p><b>Reason:</b> ${banData.details?.reason || "not specified"}</p>
        <p><b>Contacts:</b> ${banData.contact}</p>
        ${banData.details?.expires_at ? `<p><b>Block to:</b> ${new Date(banData.details.expires_at).toLocaleString()}</p>` : ''}
      </div>
    `;
    
    errorText.innerHTML = banMessage;
    errorModal.style.display = "flex";
  }

 // Function to display common errors
  function showErrorModal(message) {
    errorText.innerHTML = `<strong>Error:</strong> ${message}`;
    errorModal.style.display = "flex";
  }

 // Functions for closing modals
  window.closeErrorModal = function() {
    errorModal.style.display = "none";
  };

  // Close when clicked outside the modal
  window.addEventListener("click", function(event) {
    if (event.target === errorModal) {
      closeErrorModal();
    }
  });
});